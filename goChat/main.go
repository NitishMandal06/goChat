package main

import (
	"encoding/json"//decoding json
	"fmt"//printing to console
	"net/http"//handling http requests
	"os" //reading/writing files
	"sort"
	"strings"
	"time"
)

// User struct to store user data
type User struct {
	UserId   string `json:"userId"`
	Password string `json:"password,omitempty"` // Omit password when returning to client
	Email    string `json:"email,omitempty"`
}

// UsersData struct to match our JSON structure
type UsersData struct {
	Users []User `json:"users"`
}

// Search request struct
type SearchRequest struct {
	SearchTerm string `json:"searchTerm"`
}

// Message struct to store chat messages
type Message struct {
	Sender    string    `json:"sender"`
	Receiver  string    `json:"receiver"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	IsRead    bool      `json:"isRead"`
}

// ChatsData struct to match our JSON structure
type ChatsData struct {
	Messages []Message `json:"messages"`
}

// Message request struct
type MessageRequest struct {
	Receiver string `json:"receiver"`
	Content  string `json:"content"`
}

// RecentChat struct to store recent chat information
type RecentChat struct {
	UserId      string    `json:"userId"`
	ContactId   string    `json:"contactId"`
	LastMessage string    `json:"lastMessage"`
	Timestamp   time.Time `json:"timestamp"`
	IsRead      bool      `json:"isRead"`
}

// RecentChatsData struct to match our JSON structure
type RecentChatsData struct {
	Chats []RecentChat `json:"chats"`
}

// CORS middleware
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// Handler for user registration
func registerUser(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	
	fmt.Println("Registration request received from IP:", r.RemoteAddr)
	
	// Check if this is a browser form submission or an AJAX request
	isAjaxRequest := r.Header.Get("Content-Type") == "application/json"
	fmt.Println("Is AJAX request:", isAjaxRequest)
	
	var newUser User
	
	// Parse request differently based on request type
	if isAjaxRequest {
		// Parse JSON request body
		err := json.NewDecoder(r.Body).Decode(&newUser)
		if err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		// Parse form data
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form data: "+err.Error(), http.StatusBadRequest)
			return
		}
		
		newUser.UserId = r.FormValue("username")
		newUser.Password = r.FormValue("password")
		newUser.Email = r.FormValue("email")
		
		fmt.Println("Form data parsed - Username:", newUser.UserId, "Email:", newUser.Email)
	}
	
	// Read existing users from file
	usersData := UsersData{Users: []User{}}
	
	// Check if the file exists
	if _, err := os.Stat("users.json"); err == nil {
		// File exists, read it
		data, err := os.ReadFile("users.json")
		if err != nil {
			http.Error(w, "Error reading users.json: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Unmarshal existing user data
		err = json.Unmarshal(data, &usersData)
		if err != nil {
			http.Error(w, "Error parsing users.json: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		fmt.Println("Current number of users:", len(usersData.Users))
	} else {
		fmt.Println("users.json does not exist yet, creating new file")
	}
	
	// Check if user already exists
	for _, user := range usersData.Users {
		if user.UserId == newUser.UserId {
			fmt.Println("User already exists:", newUser.UserId)
			if isAjaxRequest {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]bool{"success": false})
			} else {
				http.Redirect(w, r, "/?error=user_exists", http.StatusFound)
			}
			return
		}
	}
	
	// Add the new user
	usersData.Users = append(usersData.Users, newUser)
	
	// Write updated users to file
	newData, err := json.MarshalIndent(usersData, "", "  ")
	if err != nil {
		http.Error(w, "Error encoding users data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	err = os.WriteFile("users.json", newData, 0644)
	if err != nil {
		http.Error(w, "Error writing to users.json: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	fmt.Println("User registered successfully:", newUser.UserId)
	
	// Return success response
	if isAjaxRequest {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	} else {
		// For form submission, redirect directly to dashboard with userId
		fmt.Println("Redirecting to dashboard after successful registration")
		http.Redirect(w, r, "/dashboard?userId="+newUser.UserId, http.StatusFound)
	}
}

// Handler for user login
func loginUser(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	
	// Check if this is a browser form submission or an AJAX request
	isAjaxRequest := r.Header.Get("Content-Type") == "application/json"
	
	var loginData User
	
	// Parse request differently based on request type
	if isAjaxRequest {
		// Parse JSON request body
		err := json.NewDecoder(r.Body).Decode(&loginData)
		if err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		// Parse form data
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form data: "+err.Error(), http.StatusBadRequest)
			return
		}
		
		loginData.UserId = r.FormValue("username")
		loginData.Password = r.FormValue("password")
	}
	
	// Read existing users from file
	usersData := UsersData{Users: []User{}}
	
	// Check if the file exists
	if _, err := os.Stat("users.json"); err != nil {
		// File doesn't exist, no users yet
		if isAjaxRequest {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]bool{"success": false})
		} else {
			http.Redirect(w, r, "/?error=invalid_credentials", http.StatusFound)
		}
		return
	}
	
	// Read the file
	data, err := os.ReadFile("users.json")
	if err != nil {
		http.Error(w, "Error reading users.json: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Unmarshal the user data
	err = json.Unmarshal(data, &usersData)
	if err != nil {
		http.Error(w, "Error parsing users.json: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Check user credentials
	for _, user := range usersData.Users {
		if user.UserId == loginData.UserId && user.Password == loginData.Password {
			// Authentication successful
			fmt.Println("User authenticated:", user.UserId)
			
			// If it's an AJAX request, return JSON response
			if isAjaxRequest {
				w.Header().Set("Content-Type", "application/json")
				
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
					"redirectTo": "/dashboard?userId=" + user.UserId,
					"userId": user.UserId,
					"email": user.Email,
				})
			} else {
				// For form submission, redirect to dashboard with userId parameter
				fmt.Println("Redirecting user to dashboard:", user.UserId)
				http.Redirect(w, r, "/dashboard?userId="+user.UserId, http.StatusFound)
			}
			return
		}
	}
	
	// If we get here, login failed
	fmt.Println("Login failed for:", loginData.UserId)
	if isAjaxRequest {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": false})
	} else {
		http.Redirect(w, r, "/?error=invalid_credentials", http.StatusFound)
	}
}

// Handler for user search
func searchUsers(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	
	// Parse request body
	var searchReq SearchRequest
	err := json.NewDecoder(r.Body).Decode(&searchReq)
	if err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Read users from file
	usersData := UsersData{Users: []User{}}
	
	// Check if the file exists
	if _, err := os.Stat("users.json"); err != nil {
		// File doesn't exist, no users yet
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"users": []User{},
		})
		return
	}
	
	// Read the file
	data, err := os.ReadFile("users.json")
	if err != nil {
		http.Error(w, "Error reading users.json: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Unmarshal the user data
	err = json.Unmarshal(data, &usersData)
	if err != nil {
		http.Error(w, "Error parsing users.json: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Filter users based on search term
	results := []User{}
	for _, user := range usersData.Users {
		// Don't include password in search results
		userWithoutPassword := User{
			UserId: user.UserId,
			Email:  user.Email,
		}
		
		// Check if user ID contains search term (case insensitive)
		if strings.Contains(strings.ToLower(user.UserId), strings.ToLower(searchReq.SearchTerm)) {
			results = append(results, userWithoutPassword)
		}
	}
	
	// Return results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"users":   results,
	})
}

// Serve the dashboard.html page
func serveDashboard(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Dashboard requested from IP:", r.RemoteAddr)
	fmt.Println("User-Agent:", r.UserAgent())
	fmt.Println("Cookies:", r.Cookies())
	http.ServeFile(w, r, "templates/dashboard.html")
}

// Serve the redirect.html page
func serveRedirect(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/redirect.html")
}

// Serve the test page
func serveTestPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/test.html")
}

// Serve static files
func setupStaticFiles() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
}

// Serve the index.html file
func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/index.html")
}

// Handler for direct to dashboard redirection
func dashboardRedirect(w http.ResponseWriter, r *http.Request) {
	// Serve an HTML page that will immediately redirect to /dashboard
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Redirecting...</title>
    <meta http-equiv="refresh" content="0;url=/dashboard">
</head>
<body>
    <p>Redirecting to dashboard...</p>
    <script>
        window.location.replace("/dashboard");
    </script>
</body>
</html>`
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// Function to update recent chats after a message is sent
func updateRecentChats(sender, receiver, message string, timestamp time.Time, isRead bool) {
	// Read existing recent chats from file
	recentChatsData := RecentChatsData{Chats: []RecentChat{}}
	
	// Check if the file exists
	if _, err := os.Stat("recentChats.json"); err == nil {
		// File exists, read it
		data, err := os.ReadFile("recentChats.json")
		if err != nil {
			fmt.Println("Error reading recentChats.json:", err)
			return
		}
		
		// Unmarshal existing recent chats data
		err = json.Unmarshal(data, &recentChatsData)
		if err != nil {
			fmt.Println("Error parsing recentChats.json:", err)
			return
		}
	}
	
	// Update recent chats for sender
	updateSingleRecentChat(&recentChatsData, sender, receiver, message, timestamp, true) // For sender, mark as read
	
	// Update recent chats for receiver
	updateSingleRecentChat(&recentChatsData, receiver, sender, message, timestamp, isRead) // For receiver, use the provided read status
	
	// Write updated recent chats to file
	newData, err := json.MarshalIndent(recentChatsData, "", "  ")
	if err != nil {
		fmt.Println("Error encoding recent chats data:", err)
		return
	}
	
	err = os.WriteFile("recentChats.json", newData, 0644)
	if err != nil {
		fmt.Println("Error writing to recentChats.json:", err)
		return
	}
	
	fmt.Println("Recent chats updated successfully")
}

// Helper function to update a single user's recent chats
func updateSingleRecentChat(data *RecentChatsData, userId, contactId, message string, timestamp time.Time, isRead bool) {
	// Check if this recent chat already exists
	found := false
	for i, chat := range data.Chats {
		if chat.UserId == userId && chat.ContactId == contactId {
			// Update existing chat
			data.Chats[i].LastMessage = message
			data.Chats[i].Timestamp = timestamp
			data.Chats[i].IsRead = isRead // Update read status
			found = true
			break
		}
	}
	
	// If not found, add a new recent chat
	if !found {
		data.Chats = append(data.Chats, RecentChat{
			UserId:      userId,
			ContactId:   contactId,
			LastMessage: message,
			Timestamp:   timestamp,
			IsRead:      isRead, // Set read status for new chat
		})
	}
}

// Handler for sending a message
func sendMessage(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	
	fmt.Println("Message send request received from IP:", r.RemoteAddr)
	
	// Get sender from request (in a real app, this would come from authentication)
	sender := r.URL.Query().Get("sender")
	if sender == "" {
		http.Error(w, "Sender is required", http.StatusBadRequest)
		return
	}
	
	// Parse request body
	var msgReq MessageRequest
	err := json.NewDecoder(r.Body).Decode(&msgReq)
	if err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Create message
	now := time.Now()
	message := Message{
		Sender:    sender,
		Receiver:  msgReq.Receiver,
		Content:   msgReq.Content,
		Timestamp: now,
		IsRead:    false, // New messages are unread by default
	}
	
	fmt.Printf("Storing message: %s -> %s: %s\n", sender, msgReq.Receiver, msgReq.Content)
	
	// Read existing chats from file
	chatsData := ChatsData{Messages: []Message{}}
	
	// Check if the file exists
	if _, err := os.Stat("chats.json"); err == nil {
		// File exists, read it
		data, err := os.ReadFile("chats.json")
		if err != nil {
			http.Error(w, "Error reading chats.json: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Unmarshal existing chat data
		err = json.Unmarshal(data, &chatsData)
		if err != nil {
			http.Error(w, "Error parsing chats.json: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		fmt.Printf("Current number of messages: %d\n", len(chatsData.Messages))
	} else {
		fmt.Println("chats.json does not exist yet, creating new file")
	}
	
	// Add the new message
	chatsData.Messages = append(chatsData.Messages, message)
	
	// Write updated chats to file
	newData, err := json.MarshalIndent(chatsData, "", "  ")
	if err != nil {
		http.Error(w, "Error encoding chats data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	err = os.WriteFile("chats.json", newData, 0644)
	if err != nil {
		http.Error(w, "Error writing to chats.json: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Update recent chats
	updateRecentChats(sender, msgReq.Receiver, msgReq.Content, now, false) // New messages are unread by default
	
	fmt.Println("Message stored successfully")
	
	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// Handler for retrieving chat messages
func getMessages(w http.ResponseWriter, r *http.Request) {
	// Only accept GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	
	// Get user IDs from query parameters
	user1 := r.URL.Query().Get("user1")
	user2 := r.URL.Query().Get("user2")
	
	if user1 == "" || user2 == "" {
		http.Error(w, "Both user IDs are required", http.StatusBadRequest)
		return
	}
	
	fmt.Printf("Retrieving chat between %s and %s\n", user1, user2)
	
	// Read chats from file
	chatsData := ChatsData{Messages: []Message{}}
	
	// Check if the file exists
	if _, err := os.Stat("chats.json"); err != nil {
		// File doesn't exist, no chats yet
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"messages": []Message{},
		})
		return
	}
	
	// Read the file
	data, err := os.ReadFile("chats.json")
	if err != nil {
		http.Error(w, "Error reading chats.json: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Unmarshal the chat data
	err = json.Unmarshal(data, &chatsData)
	if err != nil {
		http.Error(w, "Error parsing chats.json: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Filter messages between the two users
	filteredMessages := []Message{}
	for _, msg := range chatsData.Messages {
		if (msg.Sender == user1 && msg.Receiver == user2) || (msg.Sender == user2 && msg.Receiver == user1) {
			filteredMessages = append(filteredMessages, msg)
		}
	}
	
	fmt.Printf("Found %d messages between the users\n", len(filteredMessages))
	
	// Return messages
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"messages": filteredMessages,
	})
}

// Handler for retrieving all messages for a user
func getAllMessages(w http.ResponseWriter, r *http.Request) {
	// Only accept GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	
	// Get user ID from query parameter
	user := r.URL.Query().Get("user")
	
	if user == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}
	
	fmt.Printf("Retrieving all messages for user: %s\n", user)
	
	// Read chats from file
	chatsData := ChatsData{Messages: []Message{}}
	
	// Check if the file exists
	if _, err := os.Stat("chats.json"); err != nil {
		// File doesn't exist, no chats yet
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"messages": []Message{},
			"recentChats": []interface{}{},
		})
		return
	}
	
	// Read the file
	data, err := os.ReadFile("chats.json")
	if err != nil {
		http.Error(w, "Error reading chats.json: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Unmarshal the chat data
	err = json.Unmarshal(data, &chatsData)
	if err != nil {
		http.Error(w, "Error parsing chats.json: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Filter messages where the user is sender or receiver
	filteredMessages := []Message{}
	for _, msg := range chatsData.Messages {
		if msg.Sender == user || msg.Receiver == user {
			filteredMessages = append(filteredMessages, msg)
		}
	}
	
	// Create a map to aggregate recent chats
	contactsMap := make(map[string]struct {
		LastMessage string    `json:"lastMessage"`
		Timestamp   time.Time `json:"timestamp"`
	})
	
	// Get the most recent message for each contact
	for _, msg := range filteredMessages {
		var contactId string
		
		if msg.Sender == user {
			contactId = msg.Receiver
		} else {
			contactId = msg.Sender
		}
		
		// If this contact doesn't exist yet, or we have a more recent message
		if contact, exists := contactsMap[contactId]; !exists || msg.Timestamp.After(contact.Timestamp) {
			contactsMap[contactId] = struct {
				LastMessage string    `json:"lastMessage"`
				Timestamp   time.Time `json:"timestamp"`
			}{
				LastMessage: msg.Content,
				Timestamp:   msg.Timestamp,
			}
		}
	}
	
	// Convert map to array and sort by timestamp (newest first)
	type ContactInfo struct {
		UserId      string    `json:"userId"`
		LastMessage string    `json:"lastMessage"`
		Timestamp   time.Time `json:"timestamp"`
	}
	
	recentChats := []ContactInfo{}
	for userId, contact := range contactsMap {
		recentChats = append(recentChats, ContactInfo{
			UserId:      userId,
			LastMessage: contact.LastMessage,
			Timestamp:   contact.Timestamp,
		})
	}
	
	// Sort by timestamp (newest first)
	sort.Slice(recentChats, func(i, j int) bool {
		return recentChats[i].Timestamp.After(recentChats[j].Timestamp)
	})
	
	fmt.Printf("Found %d recent chats for user %s\n", len(recentChats), user)
	
	// Return messages and recent chats
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"messages":    filteredMessages,
		"recentChats": recentChats,
	})
}

// Handler for retrieving recent chats for a user
func getRecentChats(w http.ResponseWriter, r *http.Request) {
	// Only accept GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	
	// Get user ID from query parameter
	userId := r.URL.Query().Get("userId")
	
	if userId == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}
	
	fmt.Printf("Retrieving recent chats for user: %s\n", userId)
	
	// Read recent chats from file
	recentChatsData := RecentChatsData{Chats: []RecentChat{}}
	
	// Check if the file exists
	if _, err := os.Stat("recentChats.json"); err != nil {
		// File doesn't exist, no recent chats yet
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"recentChats": []RecentChat{},
		})
		return
	}
	
	// Read the file
	data, err := os.ReadFile("recentChats.json")
	if err != nil {
		http.Error(w, "Error reading recentChats.json: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Unmarshal the recent chats data
	err = json.Unmarshal(data, &recentChatsData)
	if err != nil {
		http.Error(w, "Error parsing recentChats.json: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Filter recent chats for this user
	userRecentChats := []RecentChat{}
	for _, chat := range recentChatsData.Chats {
		if chat.UserId == userId {
			userRecentChats = append(userRecentChats, chat)
		}
	}
	
	// Sort by timestamp (newest first)
	sort.Slice(userRecentChats, func(i, j int) bool {
		return userRecentChats[i].Timestamp.After(userRecentChats[j].Timestamp)
	})
	
	fmt.Printf("Found %d recent chats for user %s\n", len(userRecentChats), userId)
	
	// Return recent chats
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"recentChats": userRecentChats,
	})
}

// Handler for marking messages as read
func markMessagesAsRead(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user and contact IDs from query parameters
	userId := r.URL.Query().Get("user")
	contactId := r.URL.Query().Get("contact")

	if userId == "" || contactId == "" {
		http.Error(w, "Missing user or contact ID", http.StatusBadRequest)
		return
	}

	fmt.Printf("Marking messages from %s to %s as read\n", contactId, userId)
	
	// Read chats from file
	chatsData := ChatsData{Messages: []Message{}}
	
	// Check if the file exists
	if _, err := os.Stat("chats.json"); err != nil {
		// File doesn't exist, no messages to mark as read
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
		return
	}
	
	// Read the file
	data, err := os.ReadFile("chats.json")
	if err != nil {
		http.Error(w, "Error reading chats.json: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Unmarshal the chat data
	err = json.Unmarshal(data, &chatsData)
	if err != nil {
		http.Error(w, "Error parsing chats.json: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Mark messages as read
	messagesMarked := false
	for i, msg := range chatsData.Messages {
		// Only mark messages from the contact to the user
		if msg.Sender == contactId && msg.Receiver == userId && !msg.IsRead {
			chatsData.Messages[i].IsRead = true
			messagesMarked = true
		}
	}
	
	if messagesMarked {
		// Write updated chats to file
		newData, err := json.MarshalIndent(chatsData, "", "  ")
		if err != nil {
			http.Error(w, "Error encoding chats data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		err = os.WriteFile("chats.json", newData, 0644)
		if err != nil {
			http.Error(w, "Error writing to chats.json: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Now update the recent chats to reflect read status
		recentChatsData := RecentChatsData{Chats: []RecentChat{}}
		
		// Check if the recent chats file exists
		if _, err := os.Stat("recentChats.json"); err == nil {
			// File exists, read it
			data, err := os.ReadFile("recentChats.json")
			if err != nil {
				http.Error(w, "Error reading recentChats.json: "+err.Error(), http.StatusInternalServerError)
				return
			}
			
			// Unmarshal existing recent chats data
			err = json.Unmarshal(data, &recentChatsData)
			if err != nil {
				http.Error(w, "Error parsing recentChats.json: "+err.Error(), http.StatusInternalServerError)
				return
			}
			
			// Update the read status for the user's chat with the contact
			for i, chat := range recentChatsData.Chats {
				if chat.UserId == userId && chat.ContactId == contactId {
					recentChatsData.Chats[i].IsRead = true
					break
				}
			}
			
			// Write updated recent chats to file
			newData, err := json.MarshalIndent(recentChatsData, "", "  ")
			if err != nil {
				http.Error(w, "Error encoding recent chats data: "+err.Error(), http.StatusInternalServerError)
				return
			}
			
			err = os.WriteFile("recentChats.json", newData, 0644)
			if err != nil {
				http.Error(w, "Error writing to recentChats.json: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
		
		fmt.Println("Messages marked as read successfully")
	}
	
	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func main() {
	// Setup route handlers
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/dashboard", serveDashboard)
	http.HandleFunc("/redirect", serveRedirect)
	http.HandleFunc("/test", serveTestPage)
	http.HandleFunc("/goto-dashboard", dashboardRedirect) // New direct redirect endpoint
	http.Handle("/register", enableCORS(http.HandlerFunc(registerUser)))
	http.Handle("/login", enableCORS(http.HandlerFunc(loginUser)))
	http.Handle("/search-users", enableCORS(http.HandlerFunc(searchUsers)))
	http.Handle("/send-message", enableCORS(http.HandlerFunc(sendMessage)))
	http.Handle("/get-messages", enableCORS(http.HandlerFunc(getMessages)))
	http.Handle("/get-all-messages", enableCORS(http.HandlerFunc(getAllMessages)))
	http.Handle("/get-recent-chats", enableCORS(http.HandlerFunc(getRecentChats)))
	http.Handle("/mark-messages-read", enableCORS(http.HandlerFunc(markMessagesAsRead)))
	
	// Setup static file serving
	setupStaticFiles()
	
	// Start the server
	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
