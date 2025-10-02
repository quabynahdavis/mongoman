#!/bin/bash

currentDIR=$(pwd)
installationPATH="/usr/local/bin/mongoInstance"
sourceFile="$currentDIR/linux/mongoInstance.sh"
args1="$1"

check_mongodb_installation() {
    if type -p mongod > /dev/null; then
        echo "✅ Mongodb is installed"
    else
        echo "❗ Mongodb is not installed! Please install it and try again"
    fi
}

uninstall() {
    echo "Uninstall mongoDB instance manager"
    echo "Please wait"
    echo
    
    if ask_yes_no "Do you want to continue with uninstallation?"; then
        echo "Uninstalling"
        sudo rm -rf "$installationPATH"
        echo "😎 Finishing Up"
        if [ -f "$installationPATH" ]; then
            sudo rm  "$installationPATH"
        else
            echo "✅ Uninstallation complete"
        fi
        echo "✅ Done"
    else
        echo "❗ Uninstallation aborted! You can try again anytime"
    fi
}

checkSourceFile() {
    if [ -f "$sourceFile" ]; then
        echo "✅ Installation files are intact"
    else
        echo "❗ Installation folder is damaged or has some files missing"
        echo "😎 Reclone the repository and try again"
    fi
}

checkDestination() {
    if [ -d "$installationPATH" ]; then
        echo "✅ Installation Path exists"
    else
        echo "😎 Creating installation path"
        sudo mkdir -p "$installationPATH"
        echo "✅ Success"
    fi
}

ask_yes_no() {
    local prompt="$1"
    local default="${2:-}"  # Optional default value (y/n)
    
    while true; do
        # Show prompt with default if provided
        if [ -n "$default" ]; then
            read -p "$prompt [Y/n] " answer
        else
            read -p "$prompt [y/n] " answer
        fi
        
        # Handle empty input (use default if provided)
        if [ -z "$answer" ] && [ -n "$default" ]; then
            answer="$default"
        fi
        
        # Convert to lowercase for easier matching
        answer_lower=$(echo "$answer" | tr '[:upper:]' '[:lower:]')
        
        case "$answer_lower" in
            y|yes)
                return 0  # 0 = true/success in bash
            ;;
            n|no)
                return 1  # 1 = false/failure in bash
            ;;
            *)
                echo "Please answer with y or n"
            ;;
        esac
    done
}

startInstallation() {
    echo "😎 Installing"
    sudo cp "$sourceFile" "$installationPATH"
    echo "😎 Finishing Up"
    sudo chmod +x "$installationPATH"
    echo "✅ Done"
}

start_install_process() {
    echo "😎 Checking MongoDB installation"
    check_mongodb_installation
    
    echo
    
    echo "😎 Checking source directory"
    checkSourceFile
    
    echo
    
    echo "😎 Checking Destination directory"
    checkDestination
    
    echo
    
    echo "😎 Starting installation"
    startInstallation
    echo "✅ Installation successful!"
}

install() {
    echo "MongoInstance - mongoDB instance manager"
    echo "😎 Please wait"
    echo
    
    if ask_yes_no "Do you want to continue with installation?"; then
        echo
        clear
        start_install_process
    else
        echo "❗ Installation aborted! You can try again anytime"
    fi
}


case "${1:-}" in
    uninstall)
        uninstall
    ;;
    install)
        install
    ;;
    "")
        install
    ;;
    help)
        command...
    ;;
    *)
        command ...
    ;;
esac