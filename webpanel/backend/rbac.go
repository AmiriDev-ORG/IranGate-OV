package main
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)
type Role string
const (
	RoleAdmin     Role = "admin"
	RoleModerator Role = "moderator"
	RoleViewer    Role = "viewer"
)
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"password_hash"`
	Email        string    `json:"email"`
	Role         Role      `json:"role"`
	Permissions  []string  `json:"permissions"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    string    `json:"created_by"`
	LastLogin    time.Time `json:"last_login"`
	LastIP       string    `json:"last_ip"`
}
const (
	PermViewDashboard  = "view:dashboard"
	PermViewClients    = "view:clients"
	PermCreateClients  = "create:clients"
	PermDeleteClients  = "delete:clients"
	PermViewSettings   = "view:settings"
	PermEditSettings   = "edit:settings"
	PermViewAnalytics  = "view:analytics"
	PermManageUsers    = "manage:users"
	PermManageGroups   = "manage:groups"
	PermViewBackups    = "view:backups"
	PermCreateBackups  = "create:backups"
	PermRestoreBackups = "restore:backups"
	PermRestartServer  = "restart:server"
	PermStartServer    = "start:server"
	PermStopServer     = "stop:server"
)
var rolePermissions = map[Role][]string{
	RoleAdmin: {
		PermViewDashboard, PermViewClients, PermCreateClients, PermDeleteClients,
		PermViewSettings, PermEditSettings, PermViewAnalytics, PermManageUsers,
		PermManageGroups, PermViewBackups, PermCreateBackups, PermRestoreBackups,
		PermRestartServer, PermStartServer, PermStopServer,
	},
	RoleModerator: {
		PermViewDashboard, PermViewClients, PermCreateClients, PermDeleteClients,
		PermViewAnalytics,
	},
	RoleViewer: {
		PermViewDashboard, PermViewClients, PermViewSettings, PermViewAnalytics,
	},
}
func getDefaultPermissions(role Role) []string {
	if perms, ok := rolePermissions[role]; ok {
		return perms
	}
	return []string{}
}
func (u *User) HasPermission(permission string) bool {
	if !u.Active {
		return false
	}
	for _, p := range u.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}
var UsersDBPath = func() string {
	if dir := os.Getenv("IRANGATE_DATA_DIR"); dir != "" {
		return filepath.Join(dir, "webpanel", "users.json")
	}
	return "/opt/irangate/webpanel/users.json"
}()
func loadUsers() ([]User, error) {
	if _, err := os.Stat(UsersDBPath); os.IsNotExist(err) {
		users := []User{}
		saveUsers(users)
		return users, nil
	}
	data, err := os.ReadFile(UsersDBPath)
	if err != nil {
		return nil, err
	}
	var users []User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}
	return users, nil
}
func saveUsers(users []User) error {
	os.MkdirAll(filepath.Dir(UsersDBPath), 0755)
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(UsersDBPath, data, 0600)
}
func getUserByUsername(username string) (*User, error) {
	users, err := loadUsers()
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		if user.Username == username {
			return &user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}
func createUser(username, password, email string, role Role, createdBy string) (*User, error) {
	users, err := loadUsers()
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		if u.Username == username {
			return nil, fmt.Errorf("username already exists")
		}
	}
	user := User{
		ID:           fmt.Sprintf("user_%d", time.Now().Unix()),
		Username:     username,
		PasswordHash: hashPassword(password),
		Email:        email,
		Role:         role,
		Permissions:  getDefaultPermissions(role),
		Active:       true,
		CreatedAt:    time.Now(),
		CreatedBy:    createdBy,
	}
	users = append(users, user)
	if err := saveUsers(users); err != nil {
		return nil, err
	}
	return &user, nil
}
func updateUser(username string, updates map[string]interface{}) error {
	users, err := loadUsers()
	if err != nil {
		return err
	}
	found := false
	for i, user := range users {
		if user.Username == username {
			if email, ok := updates["email"].(string); ok {
				users[i].Email = email
			}
			if role, ok := updates["role"].(string); ok {
				users[i].Role = Role(role)
				users[i].Permissions = getDefaultPermissions(Role(role))
			}
			if active, ok := updates["active"].(bool); ok {
				users[i].Active = active
			}
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("user not found")
	}
	return saveUsers(users)
}
func deleteUser(username string) error {
	users, err := loadUsers()
	if err != nil {
		return err
	}
	var updatedUsers []User
	found := false
	for _, user := range users {
		if user.Username != username {
			updatedUsers = append(updatedUsers, user)
		} else {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("user not found")
	}
	return saveUsers(updatedUsers)
}