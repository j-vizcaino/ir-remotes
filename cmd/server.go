package cmd

import (
	"fmt"
	"github.com/j-vizcaino/ir-remotes/pkg/devices"
	"github.com/j-vizcaino/ir-remotes/pkg/remotes"
	"github.com/j-vizcaino/ir-remotes/pkg/ui"
	"github.com/j-vizcaino/ir-remotes/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cmdServer = &cobra.Command{
		Use:   "server [OPTIONS]",
		Short: "HTTP server for sending IR commands",
		Run:   Server,
	}
	listenAddress string
	frontendFile  string
)

func init() {
	flags := cmdServer.Flags()
	flags.StringVarP(&listenAddress, "listen-address", "l", ":8080", "Server listen address")
	flags.StringVar(&frontendFile, "frontend-file", "frontend.json", "JSON file with the frontend remote definitions.")

	cmdRoot.AddCommand(cmdServer)
}

type Handler struct {
	deviceInfoList devices.DeviceInfoList
	remoteList     remotes.RemoteList
}

func mustHandler() *Handler {
	devInfoList := devices.DeviceInfoList{}
	if err := utils.LoadFromFile(&devInfoList, devicesFile); err != nil {
		log.WithError(err).WithField("devices-file", devicesFile).Fatal("Failed to load devices from file.")
	}
	if len(devInfoList) == 0 {
		log.WithField("devices-file", devicesFile).Fatal("No device listed in file. Aborting.")
	}
	if err := devInfoList.InitializeDevices(udpTimeout); err != nil {
		log.WithError(err).Fatal("Failed to initialize a Broadlink device")
	}

	remoteList := remotes.RemoteList{}
	if err := utils.LoadFromFile(&remoteList, remotesFile); err != nil {
		log.WithError(err).WithField("remotes-file", remotesFile).Fatal("Failed to load remotes from file")
	}
	if len(remoteList) == 0 {
		log.WithField("remotes-file", remotesFile).Fatal("No remote listed in file. Aborting.")
	}

	return &Handler{
		deviceInfoList: devInfoList,
		remoteList:     remoteList,
	}
}

func (h *Handler) abortNotFound(c *gin.Context, err string) {
	h.abort(c, http.StatusNotFound, err)
}

func (h *Handler) abort(c *gin.Context, code int, err string) {
	c.Abort()
	c.IndentedJSON(
		code,
		gin.H{
			"success": false,
			"error":   err,
		},
	)
}

func (h *Handler) getDevices(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, h.deviceInfoList)
}

func (h *Handler) helperGetDevice(c *gin.Context, devName string) *devices.DeviceInfo {
	devInfo, found := h.deviceInfoList.Find(devices.ByName(devName))
	if !found {
		h.abortNotFound(c, fmt.Sprintf("no such device named %q", devName))
		return nil
	}
	return devInfo
}

func (h *Handler) getDevice(c *gin.Context) {
	devName := c.Param("device")

	devInfo := h.helperGetDevice(c, devName)
	if devInfo != nil {
		c.IndentedJSON(http.StatusOK, devInfo)
	}
}

func (h *Handler) getRemotes(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, h.remoteList.Names())
}

func (h *Handler) helperGetRemote(c *gin.Context) *remotes.Remote {
	name := c.Param("remote")
	remote := h.remoteList.Find(name)
	if remote == nil {
		h.abortNotFound(c, fmt.Sprintf("no such remote named %q", name))
	}
	return remote
}

func (h *Handler) getRemote(c *gin.Context) {
	r := h.helperGetRemote(c)
	if r == nil {
		return
	}
	c.IndentedJSON(http.StatusOK, r)
}

func (h *Handler) postRemoteCommand(c *gin.Context) {
	remote := h.helperGetRemote(c)
	if remote == nil {
		return
	}

	name := c.Param("command")
	cmd, ok := remote.Commands[name]
	if !ok {
		h.abortNotFound(c, fmt.Sprintf("remote %q has no command %q", remote.Name, name))
		return
	}

	// Defaults to first device unless told otherwise
	dev := h.deviceInfoList[0].GetBroadlinkDevice()

	devName, found := c.GetQuery("device")
	if found {
		devInfo := h.helperGetDevice(c, devName)
		if devInfo == nil {
			return
		}
		dev = devInfo.GetBroadlinkDevice()
	}
	if err := dev.SendIRRemoteCode(cmd, 1); err != nil {
		h.abort(c, http.StatusInternalServerError, fmt.Sprintf("IR code send failure: %s", err))
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"success": true})
}

func mustRenderIndex() string {
	remotes := make([]ui.Remote, 0)
	if err := utils.LoadFromFile(&remotes, frontendFile); err != nil {
		log.WithError(err).WithField("frontend-file", frontendFile).Fatal("Failed to load frontend data")
	}
	index, err := ui.RenderIndex(remotes)
	if err != nil {
		log.WithError(err).Fatal("Failed to render main page")
	}
	return index
}

func Server(_ *cobra.Command, _ []string) {
	h := mustHandler()

	r := gin.Default()

	indexPage := mustRenderIndex()
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Writer.WriteHeaderNow()
		c.Writer.WriteString(indexPage)
	})

	api := r.Group("/api")
	api.GET("/devices/", h.getDevices)
	api.GET("/devices/:device", h.getDevice)
	api.GET("/remotes/", h.getRemotes)
	api.GET("/remotes/:remote", h.getRemote)
	api.POST("/remotes/:remote/:command", h.postRemoteCommand)
	api.GET("/remotes/:remote/:command", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusMethodNotAllowed)
	})

	if err := r.Run(listenAddress); err != nil {
		log.WithError(err).WithField("listen-address", listenAddress).Fatal("Failed to start server")
	}
}
