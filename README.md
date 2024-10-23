# Basic System Configuration, Network Discovery, and Enumeration
This Go program acts as a helper to simplify common tasks when dealing with little known networks in Penetration Testing and Vulnerability Assessment settings.Its intuitive text-based user interface provides easy access to a variety of system and network management functions.

## Main Menu
The Main Menu of NetUtil is organized into several categories, each containing specific tools and functionalities to configure a system and analyze a network effectively.

## System Configuration
Manage and configure your system networking settings with ease.
### Check and Toggle Interfaces
View the status of all network interfaces and enable or disable them as needed.
### Edit Working Directory
Modify the directory where network captures, port scans, and analysis results are stored.
### Save Network Config
Save the current network configuration settings for future use.
### Load Network Config
Load and apply a previously saved network configuration.
## Network Recon
Perform network reconnaissance
### Wireshark Listening
Start network capturing using Wireshark and perform analysis with tshark. Serves to establish a baseline of network traffic and protocols in use, as well as identify potential areas of interest for further analysis.

## TODO
- [x] Establish Menu structure with placeholder items
- [x] Implement functions for system configuration
- [ ] Implement Discovery Scaning functionality and analysis of found systems and services
- [ ] Add Category and respective functions for Port Scanning areas of interest established during discovery scan
- [ ] TBD
