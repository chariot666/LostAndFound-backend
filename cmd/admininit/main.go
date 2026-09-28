package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"lost-found-server/internal/config"
	"lost-found-server/internal/database"
	"lost-found-server/internal/model"
)

var usernamePattern = regexp.MustCompile(`^[\p{Han}A-Za-z0-9]+$`)

func main() {
	username := flag.String("username", os.Getenv("ADMIN_USERNAME"), "admin username")
	password := flag.String("password", os.Getenv("ADMIN_PASSWORD"), "admin password")
	flag.Parse()

	if !validUsername(*username) || !validPassword(*password) {
		fmt.Fprintln(os.Stderr, "username must be 3-10 Chinese letters or digits, password must be 6-20 characters")
		os.Exit(2)
	}

	db, err := database.Open(config.Load())
	if err != nil {
		fmt.Fprintf(os.Stderr, "open database: %v\n", err)
		os.Exit(1)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hash password: %v\n", err)
		os.Exit(1)
	}

	var user model.User
	err = db.Where("username = ?", *username).First(&user).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		user = model.User{
			Username:     *username,
			PasswordHash: string(passwordHash),
			Role:         model.RoleSystemAdmin,
			Status:       model.StatusActive,
		}
		if err := db.Create(&user).Error; err != nil {
			fmt.Fprintf(os.Stderr, "create admin: %v\n", err)
			os.Exit(1)
		}
	case err != nil:
		fmt.Fprintf(os.Stderr, "find user: %v\n", err)
		os.Exit(1)
	default:
		user.PasswordHash = string(passwordHash)
		user.Role = model.RoleSystemAdmin
		user.Status = model.StatusActive
		if err := db.Save(&user).Error; err != nil {
			fmt.Fprintf(os.Stderr, "update admin: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("system admin ready: uid=%d username=%s\n", user.UID, user.Username)
}

func validUsername(username string) bool {
	length := utf8.RuneCountInString(username)
	return length >= 3 && length <= 10 && usernamePattern.MatchString(username)
}

func validPassword(password string) bool {
	length := utf8.RuneCountInString(password)
	return length >= 6 && length <= 20
}
