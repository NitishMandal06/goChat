const signInBtn = document.querySelector("#sign-in-btn");
const signUpBtn = document.querySelector("#sign-up-btn");
const container = document.querySelector(".container");
const signInForm = document.querySelector(".sign-in-form");
const signUpForm = document.querySelector(".sign-up-form");
const leftPanelContent = document.querySelector(".left-panel .content");
const rightPanelContent = document.querySelector(".right-panel .content");
const leftPanel = document.querySelector(".left-panel");
const rightPanel = document.querySelector(".right-panel");

// Disable forms during transition
function disableFormsDuringTransition() {
    document.querySelectorAll("form").forEach(form => {
        form.style.pointerEvents = "none";
        setTimeout(() => form.style.pointerEvents = "auto", 1000);
    });
}

// Reset to sign-in mode
function resetToSignInMode() {
    // Reset to sign-in mode in case URL has hash or other state
    if (container) {
        container.classList.remove("sign-up-mode");
    }
    
    // Setup initial panel visibility
    if (leftPanelContent) {
        leftPanelContent.style.opacity = "1";
        leftPanelContent.style.visibility = "visible";
    }
    
    if (rightPanelContent) {
        rightPanelContent.style.opacity = "0";
        rightPanelContent.style.visibility = "hidden";
    }
    
    // Setup pointer events
    if (leftPanel) {
        leftPanel.style.pointerEvents = "all";
    }
    
    if (rightPanel) {
        rightPanel.style.pointerEvents = "none";
    }
}

// Page load setup
document.addEventListener("DOMContentLoaded", () => {
    setTimeout(() => document.body.classList.add("loaded"), 200);
    
    // Initialize the UI
    resetToSignInMode();
    
    // Add event listeners if elements exist
    if (signInBtn) {
        signInBtn.addEventListener("click", disableFormsDuringTransition);
    }
    
    if (signUpBtn) {
        signUpBtn.addEventListener("click", disableFormsDuringTransition);
    }

    // Check if there's an error parameter in the URL
    const urlParams = new URLSearchParams(window.location.search);
    const error = urlParams.get('error');
    if (error) {
        if (error === 'invalid_credentials') {
            alert('Invalid username or password. Please try again.');
        } else if (error === 'user_exists') {
            alert('Username already exists. Please choose another username.');
            if (signUpBtn) {
                signUpBtn.click(); // Switch to signup form
            }
        }
    }
});

// Switch to Sign Up mode if the button exists
if (signUpBtn) {
    signUpBtn.addEventListener("click", () => {
        if (container) {
            container.classList.add("sign-up-mode");
        }
        
        // Hide left panel content
        if (leftPanelContent) {
            leftPanelContent.style.opacity = "0";
            leftPanelContent.style.visibility = "hidden";
        }
        
        if (leftPanel) {
            leftPanel.style.pointerEvents = "none";
        }
        
        // For mobile devices, give more time for the animation to complete
        const isMobileDevice = window.innerWidth <= 900;
        const focusDelay = isMobileDevice ? 1200 : 600;
        
        // Show right panel content with delay
        setTimeout(() => {
            if (rightPanelContent) {
                rightPanelContent.style.opacity = "1";
                rightPanelContent.style.visibility = "visible";
            }
            
            if (rightPanel) {
                rightPanel.style.pointerEvents = "all";
            }
            
            // For mobile, wait a bit longer before focusing
            setTimeout(() => {
                const usernameField = document.getElementById("signup-username");
                if (usernameField) {
                    usernameField.focus();
                }
            }, isMobileDevice ? 300 : 0);
            
        }, focusDelay);
    });
}

// Switch to Sign In mode if the button exists
if (signInBtn) {
    signInBtn.addEventListener("click", () => {
        if (container) {
            container.classList.remove("sign-up-mode");
        }
        
        // Hide right panel content
        if (rightPanelContent) {
            rightPanelContent.style.opacity = "0";
            rightPanelContent.style.visibility = "hidden";
        }
        
        if (rightPanel) {
            rightPanel.style.pointerEvents = "none";
        }
        
        // For mobile devices, give more time for the animation to complete
        const isMobileDevice = window.innerWidth <= 900;
        const focusDelay = isMobileDevice ? 1200 : 600;
        
        // Show left panel content with delay
        setTimeout(() => {
            if (leftPanelContent) {
                leftPanelContent.style.opacity = "1";
                leftPanelContent.style.visibility = "visible";
            }
            
            if (leftPanel) {
                leftPanel.style.pointerEvents = "all";
            }
            
            // For mobile, wait a bit longer before focusing
            setTimeout(() => {
                const usernameField = document.getElementById("login-username");
                if (usernameField) {
                    usernameField.focus();
                }
            }, isMobileDevice ? 300 : 0);
            
        }, focusDelay);
    });
}

// Input field animations
document.querySelectorAll(".input-field input").forEach(input => {
    input.addEventListener("focus", () => input.parentNode.classList.add("focused"));
    input.addEventListener("blur", () => {
        if(input.value === "") input.parentNode.classList.remove("focused");
    });
});

// Add a window resize listener to handle form position adjustments on resize
window.addEventListener('resize', () => {
    // Check if we're in sign-up mode
    if (container && container.classList.contains('sign-up-mode')) {
        // Reset any custom position styles
        const signInSignUp = document.querySelector('.signin-signup');
        if (signInSignUp) {
            signInSignUp.style.top = '';
            signInSignUp.style.transform = '';
        }
    } else {
        // Reset any custom position styles when in sign-in mode
        if (document.querySelector('.signin-signup')) {
            document.querySelector('.signin-signup').style.top = '';
            document.querySelector('.signin-signup').style.transform = '';
        }
    }
});