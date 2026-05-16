package utils

import (
	"log"

	"github.com/diamondburned/arikawa/v3/discord"
	"golang.org/x/crypto/bcrypt"
)

var (
	Purple discord.Color = 0x6C38C7
	Green  discord.Color = 0x27EA6B
	Red    discord.Color = 0xFF0000
	Blue   discord.Color = 0x0096FF
)

// Converts a string into a usable discord snowflake
// throws a fatal error if the snowflake is somehow invalid
func MustSnowflakeEnv(env string) discord.Snowflake {
	s, err := discord.ParseSnowflake(env)
	if err != nil {
		log.Fatalf("Invalid snowflake for $%s: %v", env, err)
	}
	return s
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// VerifyPassword verifies if the given password matches the stored hash.
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
