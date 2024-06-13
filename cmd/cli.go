package main

import (
	"atomic-pilot/config"
	"atomic-pilot/pkg/atomic"
	database "atomic-pilot/pkg/atomic/database"
	"atomic-pilot/pkg/slack"
	"atomic-pilot/utils"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	dbPath = "pkg/atomic/database/atomic_red_team.database"
	ctiUrl = "https://raw.githubusercontent.com/mitre/cti/master/enterprise-attack/enterprise-attack.json"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	confFile := flag.String("config", "config.yml", "The YAML configuration file.")
	flag.Parse()

	conf := config.Config{}
	if err := conf.Load(*confFile); err != nil {
		logger.WithError(err).WithField("config", *confFile).Fatal("failed to load configuration")
	}

	if err := conf.Validate(); err != nil {
		logger.WithError(err).WithField("config", *confFile).Fatal("invalid configuration")
	}

	logrusLevel, err := logrus.ParseLevel(conf.Log.Level)
	if err != nil {
		logger.WithError(err).Error("invalid log level provided")
		logrusLevel = logrus.InfoLevel
	}
	logger.WithField("level", logrusLevel.String()).Info("set log level")
	logger.SetLevel(logrusLevel)

	// Initialize the database
	conn, err := database.InitializeDB(dbPath)
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v\n", err)
	}
	defer database.CloseDB(conn) // Ensure the database connection is closed

	// Determine the installation path based on the OS
	installPath, err := atomic.GetInstallPath()
	if err != nil {
		logger.Fatalf("Error determining installation path: %v\n", err)
	}

	// Check if Atomic Red Team is already installed
	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		err = atomic.InstallAndVerifyAtomicRedTeam(installPath)
		if err != nil {
			logger.Fatalf("Failed to install Atomic Red Team tests: %v\n", err)
		}
	} else if err != nil {
		logger.Fatalf("Failed to check Atomic Red Team installation: %v\n", err)
	}

	// Fetch the STIX data from MITRE/CTI
	resp, err := http.Get(ctiUrl)
	if err != nil {
		logger.Fatalf("Error fetching STIX data: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalf("Failed to close body: %v\n", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Fatalf("Error reading response body: %v", err)
	}

	var stixData map[string]interface{}
	err = json.Unmarshal(body, &stixData)
	if err != nil {
		logger.Fatalf("Error unmarshalling JSON: %v", err)
	}

	objects := stixData["objects"].([]interface{})
	groups := atomic.FilterObjectsByType(objects, "intrusion-set")

	if len(groups) == 0 {
		logger.Fatal("No intrusion-set found.")
	}

	// Get the next group to process
	selectedGroup, err := atomic.GetNextGroup(groups)
	if err != nil {
		logger.Fatalf("Error selecting next group: %v", err)
	}

	groupID := selectedGroup["id"].(string)
	groupName := selectedGroup["name"].(string)
	aliases := selectedGroup["aliases"].([]interface{})
	groupName += " (AKA " + strings.Join(utils.ConvertToStringSlice(aliases), ", ") + ")"

	// Find techniques used by the group
	var techniques []string
	relationships := atomic.FilterRelationshipsBySource(objects, groupID)
	for _, relationship := range relationships {
		relationshipObj := relationship.(map[string]interface{})
		if !strings.Contains(relationshipObj["target_ref"].(string), "attack-pattern") {
			continue
		}
		technique := atomic.GetObjectByID(objects, relationshipObj["target_ref"].(string))
		if technique == nil || utils.IsDeprecatedOrRevoked(technique) {
			continue
		}
		techniqueObj := utils.ToSTIXObject(technique)
		techniqueID := techniqueObj.ExternalReferences[0].ExternalID
		techniques = append(techniques, techniqueID)
	}

	// Print the found techniques
	logger.Infof("Techniques used by %s:", groupName)
	for _, technique := range techniques {
		logger.Info(technique)
	}

	// Get all unique techniques in the database
	allTechniques, err := atomic.GetAllTechniques(objects)
	if err != nil {
		log.Fatalf("Could not get all techniques: %v\n", err)
	}

	// Check if all techniques have been used
	if atomic.AreAllTechniquesUsed(conn, allTechniques) {
		slack.SendSlackMessage("All TTPs have been used. No further actions needed.", conf.Slack.URL)
		logger.Info("All TTPs have been used. Exiting.")
		return
	}

	// Run Atomic Red Team tests for each technique
	for _, technique := range techniques {
		if atomic.HasTechniqueBeenUsed(conn, technique) {
			logger.Infof("Technique %s has already been used.", technique)
			continue
		}

		testExists, platform := atomic.CheckAtomicTestExists(technique, installPath)
		if testExists {
			slack.SendSlackMessage(fmt.Sprintf("Running Atomic Red Team test for technique %s", technique), conf.Slack.URL)
			atomic.RunAtomicTest(technique, installPath)
			atomic.MarkTechniqueAsUsed(conn, technique)
			time.Sleep(10 * time.Minute) // wait 10 minutes for the next test to be executed, so you can see if a detection was triggered
		} else {
			if platform != "" {
				slack.SendSlackMessage(fmt.Sprintf("No Atomic Red Team test available for technique %s on %s", technique, platform), conf.Slack.URL)
			} else {
				logger.Infof("No Atomic Red Team test available for technique %s", technique)
			}
		}
	}
	slack.SendSlackMessage("All tests have been executed for this run", conf.Slack.URL)
}
