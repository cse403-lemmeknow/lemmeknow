package main

import (
	webpush "github.com/Appboy/webpush-go"
)

type UserID = uint64
type GroupID = uint64
type ActivityID = uint64
type AvailabilityID = uint64
type TaskID = uint64
type UnixMillis = uint64

type Group struct {
	GroupID        GroupID `dynamo:",hash"` // Hash key, a.k.a. partition key
	Name           string
	CalendarMode   string
	Poll           *Poll
	Members        []UserID
	Activities     []Activity
	Availabilities []Availability
	Tasks          []Task
	// Counts updates to help ensure atomicity.
	UpdateCount uint64
}

type User struct {
	UserID        UserID `dynamo:",hash"` // Hash key, a.k.a. partition key
	Name          string
	Status        string
	Groups        []GroupID      `dynamo:",set"`
	Connections   []ConnectionID `dynamo:",set"`
	Subscriptions []webpush.Subscription
	// Counts updates to help ensure atomicity.
	UpdateCount uint64
}

type Poll struct {
	Title     string
	Timestamp uint64
	Options   []PollOption
	DoneFlag  bool
}

type Message struct {
	GroupID   GroupID `dynamo:",hash"`
	Timestamp uint64  `dynamo:",range"`
	Content   string
	Sender    UserID
}

type Variable struct {
	Name  string `dynamo:",hash"`
	Value string
}

type Connection struct {
	ConnectionID ConnectionID `dynamo:",hash"`
	UserID       UserID
}

type PollOption struct {
	Name  string
	Votes []UserID `dynamo:",set"`
}

type Activity struct {
	ActivityID ActivityID
	Title      string
	Date       string
	Start      string
	End        string
	Confirmed  []UserID `dynamo:",set"`
}

type Availability struct {
	AvailabilityID AvailabilityID
	UserID         UserID
	Date           string
	Start          string
	End            string
}

type Task struct {
	TaskID    TaskID
	Title     string
	Assignee  UserID
	Completed bool
}
