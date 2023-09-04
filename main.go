package main

import (
	"context"
	"encoding/json"
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
	GithubRepo     []string
	DiscordChannel string
	CustomMessage  string
	PollInterval   time.Duration
	JsonFilePath   string
}

// Discord and GitHub configurations
func readConfig() *Config {
	var cfg Config

	cfg.DiscordToken = os.Getenv("DISCORD_TOKEN")
	cfg.GithubToken = os.Getenv("GITHUB_TOKEN")

	// Read the JSON file containing repository list
	jsonFilePath := os.Getenv("JSON_FILE_PATH")
	if jsonFilePath == "" {
		jsonFilePath = "repos.json" // Default to current directory
	}

	file, err := os.ReadFile(jsonFilePath)
	if err != nil {
		log.Fatalf("Error reading .json file (%v): %v", jsonFilePath, err)
	}

	var repos struct {
		Repositories []string `json:"repositories"`
	}

	if err := json.Unmarshal(file, &repos); err != nil {
		log.Fatalf("Error unmarshaling GitHub repositories from JSON: %v", err)
	}

	// Set the GithubRepo field to the first repository in the list (you can modify this logic as needed)
	cfg.GithubRepo = repos.Repositories
	if len(cfg.GithubRepo) == 0 {
		log.Fatalf("No repositories found in the JSON file")
	}

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

	// Create a channel to signal when all Goroutines have finished
	done := make(chan bool)

	// Iterate over each repository and launch a Goroutine
	for _, repo := range cfg.GithubRepo {
		go checkRepositoryForReleases(discord, cfg, repo, githubClient, ctx, done)
	}

	// Wait for all Goroutines to finish
	for range cfg.GithubRepo {
		<-done
	}
}

func checkRepositoryForReleases(discord *discordgo.Session, cfg *Config, repo string, githubClient *github.Client, ctx context.Context, done chan<- bool) {
	var latestReleaseTag, owner, repoName string

	parts := strings.Split(repo, "/")
	if len(parts) == 2 {
		owner = parts[0]
		repoName = parts[1]
	} else {
		log.Fatalf("Invalid format. Expected 'Owner/RepoName', got '%s'", repo)
	}

	defer func() {
		done <- true // Signal that this Goroutine has finished
	}()
	for {
		// Check for new releases in the GitHub repository
		releases, _, err := githubClient.Repositories.ListReleases(ctx, owner, repoName, nil)
		if err != nil {
			log.Printf("Error fetching releases for %s: %v", repo, err)
			// Sleep and retry after an interval on error
			time.Sleep(cfg.PollInterval * time.Minute)
			continue // Continue to the next iteration
		}

		// Check if there are new releases
		if len(releases) > 0 {
			// Get the latest release
			latestRelease := releases[0]
			newReleaseTag := *latestRelease.TagName

			/// If it's a new release, send a message to Discord
			if newReleaseTag != latestReleaseTag {
				latestReleaseTag = newReleaseTag
				message := fmt.Sprintf(`# New release in https://github.com/%s/%s
> ### Version/Tag: %s
%s`, owner, repoName, latestReleaseTag, cfg.CustomMessage)
				sendMessageToDiscord(discord, cfg.DiscordChannel, message)
			}
		} else {
			// If there are no releases, send a message to Discord
			message := fmt.Sprintf("# WARNING! \n ### No releases found in https://github.com/%s/%s", cfg.GithubOwner, repo)
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
