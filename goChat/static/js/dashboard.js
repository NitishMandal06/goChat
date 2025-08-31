// DOM elements
const searchInput = document.getElementById('search-input');
const searchBtn = document.getElementById('search-btn');
const searchResults = document.getElementById('search-results');
const currentUserDisplay = document.getElementById('current-user');
const actionsMenu = document.querySelector('.actions');
const chatMessages = document.querySelector('.chat-messages');
const chatInput = document.querySelector('.chat-input-area input');
const sendBtn = document.querySelector('.send-btn');
const chatWithDisplay = document.querySelector('.chat-with p');
const contactsList = document.querySelector('.contacts-list');

// Global variables
let currentUser = null;
let currentChatUser = null;
let recentChats = [];
let lastMessageTimestamp = null;
let pollingInterval = null;
let messageCheckInterval = 8000; // 8 seconds between checks

// Initialize the dashboard
document.addEventListener('DOMContentLoaded', function() {
    console.log("Dashboard loaded");
    
    // Get username from query parameter, local storage, or server
    const urlParams = new URLSearchParams(window.location.search);
    const userId = urlParams.get('userId');
    
    if (userId) {
        // Use URL parameter
        initUser(userId);
    } else if (localStorage.getItem('currentUser')) {
        // Use localStorage
        initUser(localStorage.getItem('currentUser'));
    } else {
        // Try to get from server
        fetchFirstUser();
    }
    
    // Set up event listeners
    setupEventListeners();
});

// Initialize user session
function initUser(userId) {
    currentUser = userId;
    currentUserDisplay.textContent = currentUser;
    localStorage.setItem('currentUser', currentUser);
    console.log("User set:", currentUser);
    loadRecentChats();
    startMessagePolling();
}

// Fetch first available user if none is set
function fetchFirstUser() {
    fetch('/search-users', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ searchTerm: "" }) // Empty search returns all users
    })
    .then(response => response.json())
    .then(data => {
        if (data.success && data.users.length > 0) {
            initUser(data.users[0].userId);
        } else {
            // Redirect to login
            alert("No user logged in. Please return to login page.");
            window.location.href = '/';
        }
    })
    .catch(error => {
        console.error("Error fetching users:", error);
        alert("Error retrieving user information. Please log in again.");
        window.location.href = '/';
    });
}

// Set up all event listeners
function setupEventListeners() {
    // Logout button
    actionsMenu.addEventListener('click', function() {
        if (confirm('Do you want to logout?')) {
            stopMessagePolling();
            localStorage.removeItem('currentUser');
            window.location.href = '/';
        }
    });
    
    // Search button
    searchBtn.addEventListener('click', function() {
        const searchTerm = searchInput.value.trim();
        if (searchTerm.length < 1) {
            alert('Please enter a username to search');
            return;
        }
        searchUsers(searchTerm);
    });
    
    // Search input - Enter key
    searchInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') searchBtn.click();
    });
    
    // Send button
    sendBtn.addEventListener('click', sendMessage);
    
    // Chat input - Enter key
    chatInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            sendMessage();
        }
    });
    
    // Mobile back button
    document.querySelector('.chat-header').addEventListener('click', function() {
        if (window.innerWidth <= 768) {
            document.querySelector('.sidebar').classList.remove('inactive');
            document.querySelector('.chat-area').classList.remove('active');
        }
    });
}

// Message polling management
let isPollingLocked = false;
let pollingTimeout = null;

// Start polling for new messages
function startMessagePolling() {
    console.log("Starting message polling");
    stopMessagePolling();
    
    // First immediate check
    if (currentUser && !isPollingLocked) {
        checkForNewMessages();
    }
    
    // Then regular interval
    pollingInterval = setInterval(() => {
        if (currentUser && !isPollingLocked) {
            isPollingLocked = true;
            checkForNewMessages();
            
            // Release lock after delay
            clearTimeout(pollingTimeout);
            pollingTimeout = setTimeout(() => {
                isPollingLocked = false;
            }, 2000);
        }
    }, messageCheckInterval);
}

// Stop polling for new messages
function stopMessagePolling() {
    if (pollingInterval) {
        clearInterval(pollingInterval);
        pollingInterval = null;
        console.log("Stopped message polling");
    }
}

// Check for new messages
function checkForNewMessages() {
    if (!currentUser || window.isCheckingMessages) return;
    
    window.isCheckingMessages = true;
    console.log("Checking for new messages");
    
    // Check for messages with current chat user
    if (currentChatUser) {
        fetch(`/get-messages?user1=${currentUser}&user2=${currentChatUser}`)
            .then(response => response.json())
            .then(data => {
                if (data.success && data.messages && data.messages.length > 0) {
                    processNewMessages(data.messages);
                }
            })
            .catch(error => {
                console.error("Error checking messages:", error);
            })
            .finally(() => {
                updateRecentChats();
                window.isCheckingMessages = false;
            });
    } else {
        updateRecentChats();
        window.isCheckingMessages = false;
    }
}

// Process new messages
function processNewMessages(messages) {
    // If this is first check, just set timestamp
    if (!lastMessageTimestamp) {
        lastMessageTimestamp = new Date();
        return;
    }
    
    // Find messages newer than last check
    const newMessages = messages.filter(msg => {
        const msgTime = new Date(msg.timestamp);
        return msgTime > lastMessageTimestamp;
    });
    
    // Display new messages from others
    if (newMessages.length > 0) {
        newMessages.filter(msg => msg.sender !== currentUser)
                  .forEach(msg => displayMessage(msg));
        
        // Update timestamp to latest
        const latestMsg = messages.reduce((latest, msg) => {
            const msgTime = new Date(msg.timestamp);
            return !latest || msgTime > new Date(latest.timestamp) ? msg : latest;
        }, null);
        
        if (latestMsg) {
            lastMessageTimestamp = new Date(latestMsg.timestamp);
            chatMessages.scrollTop = chatMessages.scrollHeight;
        }
    }
}

// Update recent chats list
function updateRecentChats() {
    if (window.isUpdatingChats) return;
    
    window.isUpdatingChats = true;
    
    fetch('/get-recent-chats?userId=' + currentUser)
        .then(response => {
            if (!response.ok) {
                return fetch('/get-all-messages?user=' + currentUser);
            }
            return response.json();
        })
        .then(data => {
            if (data.success && data.recentChats) {
                handleRecentChatsUpdate(data.recentChats);
            }
        })
        .catch(error => {
            console.error("Error updating recent chats:", error);
        })
        .finally(() => {
            setTimeout(() => { window.isUpdatingChats = false; }, 500);
        });
}

// Handle recent chats update
function handleRecentChatsUpdate(serverChats) {
    let formattedChats = serverChats;
    
    // Transform to expected format if needed
    if (serverChats[0] && serverChats[0].contactId) {
        formattedChats = serverChats.map(chat => ({
            userId: chat.contactId,
            lastMessage: chat.lastMessage,
            timestamp: chat.timestamp,
            hasMessages: true
        }));
    }
    
    // Check for changes
    const previousChats = [...recentChats];
    const hasChanges = isChatsChanged(previousChats, formattedChats);
    
    // Update UI if needed
    if (hasChanges) {
        console.log("Updating chat UI due to changes");
        displayRecentChats(formattedChats);
        
        // Notify if new messages
        const hasNewMessages = formattedChats.some(chat => {
            const existingChat = previousChats.find(c => c.userId === chat.userId);
            return !existingChat || existingChat.lastMessage !== chat.lastMessage;
        });
        
        if (hasNewMessages) playNotificationSound();
    } else {
        // Just update the data
        recentChats = JSON.parse(JSON.stringify(formattedChats));
    }
}

// Check if chats lists have changed
function isChatsChanged(list1, list2) {
    if (list1.length !== list2.length) return true;
    
    return list2.some(chat => {
        const existingChat = list1.find(c => c.userId === chat.userId);
        if (!existingChat) return true;
        return chat.lastMessage !== existingChat.lastMessage;
    });
}

// Play notification sound
function playNotificationSound() {
    try {
        const audioContext = new (window.AudioContext || window.webkitAudioContext)();
        const oscillator = audioContext.createOscillator();
        const gainNode = audioContext.createGain();
        
        oscillator.type = 'sine';
        oscillator.frequency.value = 800;
        gainNode.gain.value = 0.1;
        
        oscillator.connect(gainNode);
        gainNode.connect(audioContext.destination);
        
        oscillator.start();
        oscillator.stop(audioContext.currentTime + 0.1);
    } catch (e) {
        console.error("Could not play sound:", e);
    }
}

// Load recent chats
function loadRecentChats() {
    if (!currentUser) {
        console.error("Cannot load chats without a user");
        return;
    }
    
    console.log("Loading recent chats for:", currentUser);
    
    fetch('/get-recent-chats?userId=' + currentUser)
        .then(response => {
            if (!response.ok) {
                return handleNoRecentChatsEndpoint();
            }
            return response.json();
        })
        .then(data => {
            if (data.success && data.recentChats) {
                const formattedChats = data.recentChats.map(chat => ({
                    userId: chat.contactId,
                    lastMessage: chat.lastMessage,
                    timestamp: chat.timestamp,
                    hasMessages: true
                }));
                
                displayRecentChats(formattedChats);
            }
        })
        .catch(error => {
            console.error('Error loading recent chats:', error);
        });
}

// Fallback if recent chats endpoint doesn't exist
function handleNoRecentChatsEndpoint() {
    return fetch('/get-all-messages?user=' + currentUser)
        .then(response => {
            if (!response.ok) {
                return fetchAllUsers();
            }
            return response.json();
        })
        .then(data => {
            if (data.success && data.recentChats) {
                displayRecentChats(data.recentChats);
            }
            return data;
        });
}

// Fetch all users as fallback
function fetchAllUsers() {
    return fetch('/search-users', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ searchTerm: "" })
    })
    .then(resp => {
        if (resp.ok) return resp.json();
        throw new Error("Failed to fetch users");
    })
    .then(data => {
        if (data.success) {
            fetchRecentChatsManually(data.users);
            return { success: false };
        }
        throw new Error("Failed to get users data");
    });
}

// Manually build recent chats from all messages
function fetchRecentChatsManually(users) {
    users = users.filter(user => user.userId !== currentUser);
    
    if (users.length === 0) {
        displayNoChats();
        return;
    }
    
    const chatsPromises = users.map(user => {
        return fetch(`/get-messages?user1=${currentUser}&user2=${user.userId}`)
            .then(resp => resp.json())
            .then(data => {
                if (data.success && data.messages && data.messages.length > 0) {
                    const messages = data.messages.sort((a, b) => {
                        return new Date(b.timestamp) - new Date(a.timestamp);
                    });
                    
                    return {
                        userId: user.userId,
                        lastMessage: messages[0].content,
                        timestamp: messages[0].timestamp,
                        hasMessages: true
                    };
                }
                return {
                    userId: user.userId,
                    lastMessage: "No messages yet",
                    timestamp: new Date(0),
                    hasMessages: false
                };
            });
    });
    
    Promise.all(chatsPromises)
        .then(results => {
            const chatsWithMessages = results.filter(chat => chat.hasMessages);
            
            if (chatsWithMessages.length > 0) {
                chatsWithMessages.sort((a, b) => {
                    return new Date(b.timestamp) - new Date(a.timestamp);
                });
                
                recentChats = chatsWithMessages;
                displayRecentChats(chatsWithMessages);
            } else {
                displayNoChats();
            }
        })
        .catch(error => {
            console.error('Error building recent chats:', error);
            displayNoChats();
        });
}

// Display recent chats
function displayRecentChats(chats) {
    if (!contactsList) return;
    
    // Clear existing content except heading
    const heading = contactsList.querySelector('h2');
    contactsList.innerHTML = '';
    contactsList.appendChild(heading);
    
    if (!chats || chats.length === 0) {
        displayNoChats();
        return;
    }
    
    // Store deep copy
    recentChats = JSON.parse(JSON.stringify(chats));
    
    // Add each chat to the list
    recentChats.forEach(chat => {
        // Create contact element
        const contactDiv = document.createElement('div');
        contactDiv.className = 'contact';
        contactDiv.dataset.userId = chat.userId;
        
        // First letter for avatar
        const firstLetter = chat.userId.charAt(0).toUpperCase();
        
        // Format time
        let timeDisplay = formatChatTime(chat.timestamp);
        
        // Avatar container
        const contactImgDiv = document.createElement('div');
        contactImgDiv.className = 'contact-img';
        
        // Avatar
        const avatarDiv = document.createElement('div');
        avatarDiv.className = 'avatar';
        avatarDiv.textContent = firstLetter;
        contactImgDiv.appendChild(avatarDiv);
        
        // Contact info
        const contactInfoDiv = document.createElement('div');
        contactInfoDiv.className = 'contact-info';
        
        const nameDiv = document.createElement('div');
        nameDiv.className = 'name';
        nameDiv.textContent = chat.userId;
        
        const messageDiv = document.createElement('div');
        messageDiv.className = 'message';
        messageDiv.textContent = chat.lastMessage || 'Start chatting';
        
        contactInfoDiv.appendChild(nameDiv);
        contactInfoDiv.appendChild(messageDiv);
        
        // Time info
        const messageInfoDiv = document.createElement('div');
        messageInfoDiv.className = 'message-info';
        
        const timeDiv = document.createElement('div');
        timeDiv.className = 'time';
        timeDiv.textContent = timeDisplay;
        
        messageInfoDiv.appendChild(timeDiv);
        
        // Assemble the contact
        contactDiv.appendChild(contactImgDiv);
        contactDiv.appendChild(contactInfoDiv);
        contactDiv.appendChild(messageInfoDiv);
        
        // Click event
        contactDiv.addEventListener('click', function() {
            startChatWith({ userId: chat.userId });
        });
        
        contactsList.appendChild(contactDiv);
    });
}

// Format chat timestamp
function formatChatTime(timestamp) {
    let timeDisplay = "Recently";
    if (timestamp) {
        try {
            const date = new Date(timestamp);
            if (!isNaN(date.getTime())) {
                const today = new Date();
                if (date.toDateString() === today.toDateString()) {
                    // Today, show time
                    timeDisplay = `${date.getHours()}:${date.getMinutes().toString().padStart(2, '0')}`;
                } else {
                    // Not today, show date
                    timeDisplay = `${date.getDate()}/${date.getMonth() + 1}`;
                }
            }
        } catch (e) {
            console.error("Error formatting date:", e);
        }
    }
    return timeDisplay;
}

// Display no chats message
function displayNoChats() {
    const noChatsDiv = document.createElement('div');
    noChatsDiv.className = 'no-chats';
    noChatsDiv.innerHTML = `
        <p>No chats yet.</p>
        <p>Search for users to start chatting!</p>
    `;
    
    // Clear existing content except heading
    const heading = contactsList.querySelector('h2');
    contactsList.innerHTML = '';
    contactsList.appendChild(heading);
    contactsList.appendChild(noChatsDiv);
}

// Search for users
function searchUsers(searchTerm) {
    // Show loading indicator
    searchResults.innerHTML = '<p class="loading">Searching for users...</p>';
    searchResults.style.display = 'block';
    
    // Make search request
    fetch('/search-users', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ searchTerm })
    })
    .then(response => response.json())
    .then(data => {
        searchResults.innerHTML = '';
        
        if (data.success && data.users.length > 0) {
            displaySearchResults(data.users);
        } else {
            searchResults.innerHTML = '<p>No users found matching your search.</p>';
        }
    })
    .catch(error => {
        console.error('Error searching for users:', error);
        searchResults.innerHTML = '<p>Error searching for users. Please try again.</p>';
    });
}

// Display search results
function displaySearchResults(users) {
    users.forEach(user => {
        // Skip current user
        if (user.userId === currentUser) return;
        
        // Create user item
        const userItem = document.createElement('div');
        userItem.className = 'user-item';
        userItem.dataset.userId = user.userId;
        
        // First letter for avatar
        const firstLetter = user.userId.charAt(0).toUpperCase();
        
        userItem.innerHTML = `
            <div class="avatar">${firstLetter}</div>
            <div class="user-details">
                <h4>${user.userId}</h4>
                <p>${user.email || 'No email provided'}</p>
            </div>
        `;
        
        // Add click event
        userItem.addEventListener('click', function() {
            startChatWith(user);
        });
        
        searchResults.appendChild(userItem);
    });
}

// Start chat with a user
function startChatWith(user) {
    // Set current chat user
    currentChatUser = user.userId;
    
    // Update chat header
    chatWithDisplay.textContent = user.userId;
    
    // Enable chat input
    chatInput.disabled = false;
    
    // Show loading message
    chatMessages.innerHTML = `
        <div class="message system">
            <div class="content">
                <p>Loading messages with ${user.userId}...</p>
            </div>
        </div>
    `;
    
    // Reset last message timestamp
    lastMessageTimestamp = null;
    
    // Hide search results
    searchResults.style.display = 'none';
    
    // On mobile, switch to chat view
    if (window.innerWidth <= 768) {
        document.querySelector('.sidebar').classList.add('inactive');
        document.querySelector('.chat-area').classList.add('active');
    }
    
    // Load messages
    loadMessages(currentUser, currentChatUser);
}

// Load messages between users
function loadMessages(user1, user2) {
    if (!user1 || !user2) {
        console.error("Cannot load messages, missing user IDs:", { user1, user2 });
        chatMessages.innerHTML = `
            <div class="message system">
                <div class="content">
                    <p>Error: Cannot load messages without valid user IDs.</p>
                </div>
            </div>
        `;
        return;
    }

    console.log(`Loading messages between ${user1} and ${user2}`);
    
    fetch(`/get-messages?user1=${user1}&user2=${user2}`)
        .then(response => response.json())
        .then(data => {
            // Clear chat area
            chatMessages.innerHTML = '';
            
            if (data.success) {
                if (data.messages.length === 0) {
                    // No messages yet
                    chatMessages.innerHTML = `
                        <div class="message system">
                            <div class="content">
                                <p>This is the beginning of your conversation with ${currentChatUser}</p>
                                <div class="time">Just now</div>
                            </div>
                        </div>
                    `;
                } else {
                    // Display messages
                    data.messages.forEach(msg => displayMessage(msg));
                    
                    // Store timestamp and scroll
                    const latestMsg = data.messages.reduce((latest, msg) => {
                        const msgTime = new Date(msg.timestamp);
                        return !latest || msgTime > new Date(latest.timestamp) ? msg : latest;
                    }, null);
                    
                    if (latestMsg) {
                        lastMessageTimestamp = new Date(latestMsg.timestamp);
                        chatMessages.scrollTop = chatMessages.scrollHeight;
                    }
                }
            } else {
                chatMessages.innerHTML = `
                    <div class="message system">
                        <div class="content">
                            <p>Error loading messages. Please try again.</p>
                        </div>
                    </div>
                `;
            }
        })
        .catch(error => {
            console.error('Error loading messages:', error);
            chatMessages.innerHTML = `
                <div class="message system">
                    <div class="content">
                        <p>Error loading messages. Please try again.</p>
                    </div>
                </div>
            `;
        });
}

// Display a message
function displayMessage(message) {
    const messageEl = document.createElement('div');
    const isSent = message.sender === currentUser;
    
    messageEl.className = `message ${isSent ? 'sent' : 'received'}`;
    
    // Format timestamp
    let timeStr = formatMessageTime(message.timestamp);
    
    messageEl.innerHTML = `
        <div class="content">
            <p>${message.content}</p>
            <div class="time">${timeStr}</div>
        </div>
    `;
    
    chatMessages.appendChild(messageEl);
}

// Format message timestamp
function formatMessageTime(timestamp) {
    let timeStr = "Just now";
    try {
        const date = new Date(timestamp);
        if (!isNaN(date.getTime())) {
            timeStr = `${date.getHours()}:${date.getMinutes().toString().padStart(2, '0')}`;
        }
    } catch (e) {
        console.error("Error formatting date:", e);
    }
    return timeStr;
}

// Send message
function sendMessage() {
    const message = chatInput.value.trim();
    
    if (!message || !currentChatUser || chatInput.disabled) return;
    
    if (!currentUser) {
        alert("You must be logged in to send messages.");
        return;
    }
    
    console.log(`Sending message from ${currentUser} to ${currentChatUser}: ${message}`);
    
    // Create message object
    const msgObj = {
        receiver: currentChatUser,
        content: message
    };
    
    // Create and display message locally
    const now = new Date();
    displayMessage({
        sender: currentUser,
        receiver: currentChatUser,
        content: message,
        timestamp: now
    });
    
    // Clear input
    chatInput.value = '';
    
    // Scroll to bottom
    chatMessages.scrollTop = chatMessages.scrollHeight;
    
    // Update last message timestamp
    lastMessageTimestamp = now;
    
    // Send to server
    fetch(`/send-message?sender=${currentUser}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(msgObj)
    })
    .then(response => response.json())
    .then(data => {
        if (!data.success) {
            alert('Failed to send message. Please try again.');
        }
    })
    .catch(error => {
        console.error('Error sending message:', error);
        alert('Failed to send message. Please try again.');
    });
} 