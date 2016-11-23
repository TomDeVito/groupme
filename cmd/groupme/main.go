package main

import (
	"bufio"
	"fmt"
	"os"
	"text/template"

	"github.com/jlubawy/groupme"
)

var app *groupme.App
var token = os.Getenv("GROUPME_TOKEN")

var commands = []struct {
	Name string
	Fxn  func()
	Help string
}{
	{"group", cmdGroup, "get information about a specific group"},
	{"groups", cmdGroups, "list groups for a given user"},
	{"me", cmdUserMe, "get info about your user account"},
	{"message", cmdMessage, "send message to a group"},
	{"messages", cmdMessages, "list messages for group given index"},
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	if os.Args[1] == "help" || token == "" {
		usage()
	}

	app = groupme.NewApp(token)

	for _, cmd := range commands {
		if cmd.Name == os.Args[1] {
			cmd.Fxn()
		}
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `GroupMe is a tool for interacting with the GroupMe API.

Usage:

    groupme command [arguments]

Must set environment variable:

    GROUPME_TOKEN=%s

The commands are:

`, token)

	for _, cmd := range commands {
		fmt.Fprintf(os.Stderr, "    %-10s    %s\n", cmd.Name, cmd.Help)
	}

	fmt.Fprintf(os.Stderr, "\n")

	os.Exit(1)
}

func cmdGroup() {
	if len(os.Args) < 3 {
		fatalf("must provide group ID\n")
	}

	group := getGroup(os.Args[2])

	if err := templGroup.Execute(os.Stdout, group); err != nil {
		panic(err)
	}
}

var templGroup = template.Must(template.New("").Parse(`
Id            : {{.Id}}
GroupId       : {{.GroupId}}
Name          : {{.Name}}
PhoneNumber   : {{.PhoneNumber}}
Type          : {{.Type}}
Description   : {{.Description}}
ImageUrl      : {{.ImageUrl}}
CreatorUserId : {{.CreatorUserId}}
CreatedAt     : {{.CreatedAt}}
UpdatedAt     : {{.UpdatedAt}}
OfficeMode    : {{.OfficeMode}}
ShareUrl      : {{.ShareUrl}}
Members       : {{range .Members}}
    * Id          : {{.Id}}
      UserId      : {{.UserId}}
      Name        : {{.Name}}
      Nickname    : {{.Nickname}}
      PhoneNumber : {{.PhoneNumber}}
      Email       : {{.Email}}
      Sms         : {{.Sms}}
      ImageUrl    : {{.ImageUrl}}
      Muted       : {{.Muted}}
      Autokicked  : {{.Autokicked}}
      CreatedAt   : {{.CreatedAt}}
      UpdatedAt   : {{.UpdatedAt}}
{{end}}
Messages      : {{.Messages}}
MaxMembers    : {{.MaxMembers}}
`))

func cmdGroups() {
	groups, err := app.Groups()
	if err != nil {
		fatalf("%s", err)
	}

	for _, group := range groups {
		fmt.Printf("  %8s - %s\n", group.Id, group.Name)
	}
}

func cmdUserMe() {
	me, err := app.GetUserMe()
	if err != nil {
		fatalf("%s", err)
	}

	fmt.Printf(`{
  "id": "%s",
  "phone_number": "%s",
  "image_url": "%s",
  "name": "%s",
  "created_at": %d,
  "updated_at": %d,
  "email": "%s",
  "sms": %t
}`, me.Id, me.PhoneNumber, me.ImageUrl, me.Name, me.CreatedAt, me.UpdatedAt, me.Email, me.Sms)
}

func cmdMessage() {
	if len(os.Args) < 3 {
		fatalf("must provide group ID\n")
	}

	group := getGroup(os.Args[2])

	fmt.Printf("> ")

	line, _, _ := bufio.NewReader(os.Stdin).ReadLine()
	if line != nil {
		msg, err := app.SendMessageText(group, string(line))
		if err != nil {
			fatalf("%s", err)
		}

		fmt.Printf("\nSent message '%s' to group '%s'\n", msg.Text, group.Name)
	}
}

func cmdMessages() {
	if len(os.Args) < 3 {
		fatalf("must provide group ID\n")
	}

	group := getGroup(os.Args[2])

	msgs, err := app.GetMessages(group, 100)
	if err != nil {
		fatalf("%s", err)
	}

	for _, msg := range msgs {
		fmt.Printf("%s [system %t] %-16s (%s)  '%s'  %v\n", msg.Id, msg.System, msg.Name, msg.UserId, msg.Text, msg.Attachments)
	}
}

func getGroup(id string) *groupme.Group {
	group, err := app.Group(id)
	if err != nil {
		fatalf("%s", err)
	}

	if group == nil {
		fatalf("could not find group with ID %s\n", id)
	}

	return group
}

func fatalf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
	os.Exit(1)
}
