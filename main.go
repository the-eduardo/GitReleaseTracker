package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
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
	PollInterval   time.Duration
}

// Discord and GitHub configurations
func readConfig() *Config {
	var cfg Config

	cfg.DiscordToken = os.Getenv("DISCORD_TOKEN")
	cfg.GithubToken = os.Getenv("GITHUB_TOKEN")
	cfg.GithubOwner = os.Getenv("GITHUB_OWNER")
	cfg.GithubRepo = os.Getenv("GITHUB_REPO")
	cfg.DiscordChannel = os.Getenv("DISCORD_CHANNEL")

	// Parse poll interval from environment variable (in minutes)
	pollIntervalStr := os.Getenv("WAITING_TIME")
	PollInterval, err := strconv.Atoi(pollIntervalStr)
	if err != nil {
		log.Fatalf("failed to parse WAITING_TIME: %v", err)
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
		"GITHUB_OWNER",
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
				message := fmt.Sprintf("# New release in https://github.com/%s/%s\n> ### Version/Tag: %s", cfg.GithubOwner, cfg.GithubRepo, latestReleaseTag)
				sendMessageToDiscord(discord, cfg.DiscordChannel, message)
			}
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
