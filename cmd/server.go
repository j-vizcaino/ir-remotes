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

func mustLoadDevices() (devices.DeviceInfoList, []broadlink.Device) {
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
	return devInfoList, devList
}

func mustLoadRemotes() remotes.RemoteList {
	remotes := remotes.RemoteList{}
	if err := utils.LoadFromFile(&remotes, remotesFile); err != nil {
		log.WithError(err).WithField("remotes-file", remotesFile).Fatal("Failed to load remotes from file")
	}
	if len(remotes) == 0 {
		log.WithField("remotes-file", remotesFile).Fatal("No remote listed in file. Aborting.")
	}
	return remotes
}

func Server(_ *cobra.Command, _ []string) {
	devInfoList, deviceList := mustLoadDevices()
	remoteList := mustLoadRemotes()

	r := gin.Default()

	api := r.Group("/api")
	api.GET("/devices/", func(c *gin.Context) {
		c.JSON(http.StatusOK, devInfoList)
	})

	api.GET("/devices/:device", func(c *gin.Context) {
		devName := c.Param("device")
		devInfo, found := devInfoList.Find(func(d devices.DeviceInfo) bool {
			return d.Name == devName
		})
		if !found {
			c.AbortWithStatusJSON(
				http.StatusNotFound,
				gin.H{
					"success": false,
					"error": fmt.Sprintf("no such device named %q", devName),
				},
			)
			return
		}
		c.JSON(http.StatusOK, devInfo)
	})

	api.GET("/remotes/", func(c *gin.Context) {
		c.JSON(http.StatusOK, remoteList.Names())
	})

	api.GET("/remotes/:remote", func(c *gin.Context) {
		name := c.Param("remote")
		remote := remoteList.Find(name)
		if remote == nil {
			c.AbortWithStatusJSON(
				http.StatusNotFound,
				gin.H{
					"success": false,
					"error": fmt.Sprintf("no such remote named %q", name),
				},
			)
			return
		}
		c.JSON(http.StatusOK, remote)
	})

	api.GET("/remotes/:remote/:command", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusMethodNotAllowed)
	})
	api.POST("/remotes/:remote/:command", func(c *gin.Context) {
		name := c.Param("remote")
		remote := remoteList.Find(name)
		if remote == nil {
			c.AbortWithStatusJSON(
				http.StatusNotFound,
				gin.H{
					"success": false,
					"error": fmt.Sprintf("no such remote named %q", name),
				},
			)
			return
		}
		name = c.Param("command")
		cmd, ok := remote.Commands[name]
		if !ok {
			c.AbortWithStatusJSON(
				http.StatusNotFound,
				gin.H{
					"success": false,
					"error": fmt.Sprintf("remote %q has no command %q", remote.Name, name),
				},
			)
			return
		}
		dev := deviceList[0]
		if err := dev.SendIRRemoteCode(cmd, 1); err != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{
					"success": false,
					"error": fmt.Sprintf("send IR command failed: %s", err),
				},
			)
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	if err := r.Run(listenAddress); err != nil {
		log.WithError(err).WithField("listen-address", listenAddress).Fatal("Failed to start server")
	}
}
