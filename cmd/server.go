package cmd

import (
	"fmt"
	"net/http"

	"github.com/j-vizcaino/ir-remotes/pkg/devices"
	"github.com/j-vizcaino/ir-remotes/pkg/remotes"
	"github.com/j-vizcaino/ir-remotes/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/mixcode/broadlink"
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
)

func init() {
	cmdServer.Flags().StringVarP(&listenAddress, "listen-address", "l", ":8080", "Server listen address")
	cmdRoot.AddCommand(cmdServer)
}

type Handler struct {
	deviceInfoList devices.DeviceInfoList
	deviceList []broadlink.Device
	remoteList remotes.RemoteList
}

func mustHandler() *Handler {
	devInfoList := devices.DeviceInfoList{}
	if err := utils.LoadFromFile(&devInfoList, devicesFile); err != nil {
		log.WithError(err).WithField("devices-file", devicesFile).Fatal("Failed to load devices from file.")
	}
	if len(devInfoList) == 0 {
		log.WithField("devices-file", devicesFile).Fatal("No device listed in file. Aborting.")
	}
	devList, err := devInfoList.Initialize()
	if err != nil {
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
		deviceList: devList,
		remoteList: remoteList,
	}
}

func (h *Handler) abortNotFound(c *gin.Context, err string) {
	h.abort(c, http.StatusNotFound, err)
}

func (h *Handler) abort(c *gin.Context, code int, err string) {
	c.AbortWithStatusJSON(
		code,
		gin.H{
			"success": false,
			"error": err,
		},
	)
}

func (h *Handler) getDevices(c *gin.Context) {
	c.JSON(http.StatusOK, h.deviceInfoList)
}

func (h *Handler) getDevice(c *gin.Context) {
	devName := c.Param("device")
	devInfo, found := h.deviceInfoList.Find(func(d devices.DeviceInfo) bool {
		return d.Name == devName
	})
	if !found {
		h.abortNotFound(c, fmt.Sprintf("no such device named %q", devName))
		return
	}
	c.JSON(http.StatusOK, devInfo)
}

func (h *Handler) getRemotes(c *gin.Context) {
	c.JSON(http.StatusOK, h.remoteList.Names())
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
	c.JSON(http.StatusOK, r)
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
	dev := h.deviceList[0]
	if err := dev.SendIRRemoteCode(cmd, 1); err != nil {
		h.abort(c, http.StatusInternalServerError, fmt.Sprintf("send IR command failed: %s", err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}


func Server(_ *cobra.Command, _ []string) {
	h := mustHandler()

	r := gin.Default()

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
