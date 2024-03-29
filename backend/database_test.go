package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func isInCI() bool {
	return os.Getenv("GITHUB_ACTIONS") != ""
}

func maybeSkip(t *testing.T) {
	if !isInCI() {
		t.Skip("skipping outside of CI")
	}
}

func TestGenerateID(t *testing.T) {
	generated := make(map[uint64]struct{})
	for i := 0; i < 1000; i++ {
		id := GenerateID()

		if _, ok := generated[id]; ok {
			t.Fail()
		}
		generated[id] = struct{}{}

		assert.Positive(t, id)

		double := float64(id)
		roundtrip := uint64(double)
		assert.Equal(t, id, roundtrip)
	}
}

func TestInsertUser(t *testing.T) {
	maybeSkip(t)
	//Insert user
	table := NewDynamoDB(nil)

	var userID UserID = GenerateID()

	item := User{
		UserID:      userID,
		Name:        "Bob",
		Groups:      []GroupID{},
		Connections: []ConnectionID{},
	}

	err := table.CreateUser(item)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	////assert.Equal(t, 1, len(table.users.primarykeys()))
}

func TestDeleteUser(t *testing.T) {
	maybeSkip(t)
	//Insert user
	table := NewDynamoDB(nil)

	var userID UserID = GenerateID()

	item := User{
		UserID:      userID,
		Name:        "Bob",
		Groups:      []GroupID{},
		Connections: []ConnectionID{},
	}

	err := table.CreateUser(item)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	//assert.Equal(t, 1, len(table.users.primarykeys()))

	err = table.DeleteUser(userID)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	//assert.Equal(t, 0, len(table.users.primarykeys()))
}

func TestReadUser(t *testing.T) {
	maybeSkip(t)
	//Insert user
	table := NewDynamoDB(nil)

	var userID UserID = GenerateID()

	item := User{
		UserID:      userID,
		Name:        "Bob",
		Groups:      []GroupID{},
		Connections: []ConnectionID{},
	}

	err := table.CreateUser(item)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	//assert.Equal(t, 1, len(table.users.primarykeys()))

	user, err := table.ReadUser(userID)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	assert.Equal(t, "Bob", user.Name)
}

func TestInsertGroup(t *testing.T) {
	maybeSkip(t)
	//Insert user
	table := NewDynamoDB(nil)

	var groupID GroupID = GenerateID()

	item := Group{
		GroupID: groupID,
		Name:    "Portland",
		Poll:    nil,
	}

	err := table.CreateGroup(item)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	//assert.Equal(t, 1, len(table.users.primarykeys()))
}

func TestDeleteGroup(t *testing.T) {
	maybeSkip(t)
	//Insert user
	table := NewDynamoDB(nil)

	var groupID GroupID = GenerateID()

	item := Group{
		GroupID: groupID,
		Name:    "Portland",
		Poll:    nil,
	}

	err := table.CreateGroup(item)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	//assert.Equal(t, 1, len(table.users.primarykeys()))

	err = table.DeleteGroup(groupID)
	if err != nil {
		t.Error("unexpected error:", err)
	}
}

func TestReadGroup(t *testing.T) {
	maybeSkip(t)
	//Insert user
	table := NewDynamoDB(nil)

	var GroupID uint64 = 43

	item := Group{
		GroupID: GroupID,
		Name:    "Portland",
		Poll:    nil,
	}

	err := table.CreateGroup(item)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	//assert.Equal(t, 1, len(table.users.primarykeys()))

	group, err := table.ReadGroup(GroupID)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	assert.Equal(t, "Portland", group.Name)
}
