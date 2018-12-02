package main

import (
	"fmt"
	"github.com/j-vizcaino/ir-remotes/pkg/commands"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/mixcode/broadlink"
)

type RemoteRegistry map[string]commands.CommandRegistry

func (r RemoteRegistry) Remotes() []string {
	out := make([]string, 0, len(r))
	for k := range r {
		out = append(out, k)
	}
	return out
}

func findDevice(timeout time.Duration) broadlink.Device {
	log.Info("Looking for Broadlink devices on your network. Please wait...")
	devs, err := broadlink.DiscoverDevices(timeout, 0)
	if err != nil {
		log.WithError(err).Fatal("Failed to discover Broadlink devices")
	}
	if len(devs) == 0 {
		log.Fatal("No Broadlink device found")
	}

	// TODO: add code to pick the right one, in case multiple devices are detected
	d := devs[0]
	log.WithField("address", d.UDPAddr.String()).
		WithField("mac", fmt.Sprintf("%02x", d.MACAddr)).
		Info("Device found!")

	myname, err := os.Hostname() // Your local machine's name.
	if err != nil {
		log.WithError(err).Fatal("Failed to get hostname")
	}
	myid := make([]byte, 15)   // Must be 15 bytes long.

	err = d.Auth(myid, myname) // d.ID and d.AESKey will be updated on success.
	if err != nil {
		log.WithError(err).Fatal("Failed to authenticate with device", "addr", d.UDPAddr.String(), "mac", d.MACAddr)
	}
	return d
}

func handleRemotes(reg RemoteRegistry) func(*gin.Context) {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, reg.Remotes())
	}
}

func handleRemoteSendCommand(reg RemoteRegistry, bd broadlink.Device) func(*gin.Context) {
	return func(c *gin.Context) {
		remoteName := c.Param("remote")
		remote, ok := reg[remoteName]
		if !ok {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "no such remote " + remoteName})
			return
		}

		commandName := c.Param("command")
		command, ok := remote[commandName]
		if !ok {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("no such command %s for remote %s", commandName, remoteName)})
			return
		}

		if err := bd.SendIRRemoteCode(command, 1); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to send remote %s command %s: %s", remoteName, commandName, err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func handleRemoteCommands(reg RemoteRegistry) func(*gin.Context) {
	return func(c *gin.Context) {
		remoteName := c.Param("remote")
		remote, ok := reg[remoteName]
		if !ok {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "no such remote " + remoteName})
			return
		}
		c.JSON(http.StatusOK, remote.Commands())
	}
}

func handleDevice(bd broadlink.Device) func(*gin.Context) {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"address": bd.UDPAddr.String(),
			"mac_address": fmt.Sprintf("%x", bd.MACAddr),
		})
	}
}

func main() {
	r := gin.Default()

	bd := findDevice(time.Second)

	cmdReg := commands.CommandRegistry{}
	cmdReg.LoadFromFile("amp.json")
	remotes := RemoteRegistry{}
	remotes["amp"] = cmdReg

	r.GET("/api/device", handleDevice(bd))
	r.GET("/api/remotes/", handleRemotes(remotes))
	r.GET("/api/remotes/:remote/commands/", handleRemoteCommands(remotes))
	r.POST("/api/remotes/:remote/commands/:command", handleRemoteSendCommand(remotes, bd))
	r.Run(":12345") // listen and serve on 0.0.0.0:8080
}