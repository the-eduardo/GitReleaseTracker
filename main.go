package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/go-github/v38/github"
	"golang.org/x/oauth2"
)

type Config struct {
	DiscordToken   string
	GithubToken    string
	GithubOwner    string
	GithubRepo     string
	DiscordChannel string
	CustomMessage  string
	PollInterval   time.Duration
}

// Discord and GitHub configurations
func readConfig() *Config {
	var cfg Config

	cfg.DiscordToken = os.Getenv("DISCORD_TOKEN")
	cfg.GithubToken = os.Getenv("GITHUB_TOKEN")
	inputURL := os.Getenv("GITHUB_REPO")
	// Split the URL by "/"
	parts := strings.Split(inputURL, "/")
	// Check if the URL has the expected number of parts
	if len(parts) < 2 {
		log.Fatalf("Invalid GitHub repo! Try something like -the-eduardo/GitReleaseTracker- your input: %s", inputURL)
	}
	cfg.GithubOwner = parts[len(parts)-2]
	cfg.GithubRepo = parts[len(parts)-1]

	cfg.DiscordChannel = os.Getenv("DISCORD_CHANNEL")
	cfg.CustomMessage = os.Getenv("CUSTOM_DISCORD_MESSAGE")

	// Parse poll interval from environment variable (in minutes)
	pollIntervalStr := os.Getenv("WAITING_TIME")
	PollInterval, err := strconv.Atoi(pollIntervalStr)
	if err != nil {
		PollInterval = 60
	}
	cfg.PollInterval = time.Duration(PollInterval)

	// Set default values if not provided (60 minutes)
	if cfg.PollInterval < 5 || cfg.PollInterval > 1440 {
		log.Printf("%v is invalid WAITING_TIME, using default value (60 minutes)", cfg.PollInterval)
		cfg.PollInterval = 60
	}

	// Check required environment variables
	requiredEnvVars := []string{
		"DISCORD_TOKEN",
		"GITHUB_TOKEN",
		"GITHUB_REPO",
		"DISCORD_CHANNEL",
	}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("%s environment variable not set", envVar)
		}
	}

	return &cfg
}

func main() {
	// Read the configuration
	cfg := readConfig()

	// Create a new Discord session
	discord, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}
	defer discord.Close()

	// Create a GitHub client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GithubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	githubClient := github.NewClient(tc)

	// Keep track of the latest release tag
	var latestReleaseTag string

	for {
		// Check for new releases in the GitHub repository
		releases, _, err := githubClient.Repositories.ListReleases(ctx, cfg.GithubOwner, cfg.GithubRepo, nil)
		if err != nil {
			log.Printf("Error fetching releases: %v", err)
			continue
		}

		// Check if there are new releases
		if len(releases) > 0 {
			// Get the latest release
			latestRelease := releases[0]
			newReleaseTag := *latestRelease.TagName

			// If it's a new release, send a message to Discord
			if newReleaseTag != latestReleaseTag {
				latestReleaseTag = newReleaseTag
				message := fmt.Sprintf(`# New release in https://github.com/%s/%s\n> ### Version/Tag: %s\n%s`, cfg.GithubOwner, cfg.GithubRepo, latestReleaseTag, cfg.CustomMessage)
				sendMessageToDiscord(discord, cfg.DiscordChannel, message)
			}
		} else {
			// If there are no releases, send a message to Discord
			message := fmt.Sprintf("# WARNING! \n ### No releases found in https://github.com/%s/%s", cfg.GithubOwner, cfg.GithubRepo)
			sendMessageToDiscord(discord, cfg.DiscordChannel, message)
		}

		// Sleep for the defined interval before checking again
		time.Sleep(cfg.PollInterval * time.Minute)
	}
}

func sendMessageToDiscord(session *discordgo.Session, channelID, message string) {
	_, err := session.ChannelMessageSend(channelID, message)
	if err != nil {
		log.Printf("Error sending message to Discord: %v", err)
	}
}
