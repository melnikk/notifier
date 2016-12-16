package notifier

import (
	"fmt"
	"os"
	"strings"

	"github.com/garyburd/redigo/redis"
)

// GetIDByUsername read ID of user by messenger username
func (connector *DbConnector) GetIDByUsername(messenger, username string) (string, error) {
	if strings.HasPrefix(username, "#") {
		result := "@" + username[1:]
		log.Debugf("Channel %s requested. Returning id: %s", username, result)
		return result, nil
	}

	c := connector.Pool.Get()
	defer c.Close()

	result, err := redis.String(c.Do("GET", fmt.Sprintf("moira-%s-users:%s", messenger, username)))

	return result, err
}

// SetUsernameID store id of username
func (connector *DbConnector) SetUsernameID(messenger, username, id string) error {
	c := connector.Pool.Get()
	defer c.Close()
	if _, err := c.Do("SET", usernameKey(messenger, username), id); err != nil {
		return err
	}
	return nil
}

func usernameKey(messenger, username string) string {
	return fmt.Sprintf("moira-%s-users:%s", messenger, username)
}

const (
	botUsername  = "moira-bot-host"
	deregistered = "deregistered"
)

var messengers = make(map[string]bool)

// RegisterBotIfAlreadyNot creates registration of bot instance in redis
func (connector *DbConnector) RegisterBotIfAlreadyNot(messenger string) bool {
	host, _ := os.Hostname()
	redisKey := usernameKey(messenger, botUsername)
	c := connector.Pool.Get()
	defer c.Close()

	c.Send("WATCH", redisKey)

	status, err := redis.Bytes(c.Do("GET", redisKey))
	statusStr := string(status)
	if err != nil {
		log.Info(err)
	}
	if statusStr == "" || statusStr == host || statusStr == deregistered {
		c.Send("MULTI")
		c.Send("SET", redisKey, host)
		_, err := c.Do("EXEC")
		if err != nil {
			log.Info(err)
			return false
		}
		messengers[messenger] = true
		return true
	}

	return false
}

// DeregisterBots cancels registration for all registered messengers
func (connector *DbConnector) DeregisterBots() {
	for messenger, flag := range messengers {
		if flag {
			connector.DeregisterBot(messenger)
		}
	}
}

// DeregisterBot removes registration of bot instance in redis
func (connector *DbConnector) DeregisterBot(messenger string) error {
	status, _ := connector.GetIDByUsername(messenger, botUsername)
	host, _ := os.Hostname()
	if status == host {
		log.Debugf("Bot for %s on host %s exists. Removing registration.", messenger, host)
		return connector.SetUsernameID(messenger, botUsername, deregistered)
	}

	log.Debugf("Notifier on host %s did't exist. Removing skipped.", host)
	return nil
}
