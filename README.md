# Basic System Configuration, Network Discovery, and Enumeration
This Go program acts as a helper to simplify common tasks when dealing with little known networks in Penetration Testing and Vulnerability Assessment settings. Its intuitive text-based user interface provides easy access to a variety of system and network management functions.

## TODO
- [x] Establish Menu structure
- [x] Create Management for floating I/O Boxes
- [x] Implement functions for inital system configuration
- [x] Add Wireshark Listening and tshark analysis
- [ ] Add further system configuration based on Listening analysis
- [x] Implement Discovery Scanning functionality
- [ ] Implement analysis of Discovery Scan results
- [ ] Implement detailed Port Scanning of areas of interest
- [ ] \(Optional) Add functionality to gather configuration of network devices via SSH
- [ ] TBD

## Demo Proof of Concept
![](https://github.com/fortifyde/netutil/blob/master/demo.gif)

## Prerequisites
### Required Software
Install the Go programming language from [https://go.dev/doc/install](https://go.dev/doc/install) and use your preferred package manager to install the following:
- Wireshark & tshark (comes with Wireshark as standard on most Linux distributions)
- Zenity
- nmap
- TBD

```bash
sudo apt-get install wireshark zenity nmap
```
### Root Access
  Several functionalities require root access (sudo) to operate correctly:
  - Network interface management
  - Wireshark packet capture
  - Network configuration changes
  - Port Scanning

## Main Menu
The Main Menu of NetUtil is organized into several categories, each containing specific tools and functionalities to configure a system and analyze a network effectively.

## System Configuration
Manage and configure your system networking settings with ease.
#### Check and Toggle Interfaces
View the status of all network interfaces and enable or disable them as needed.
#### Edit Working Directory
Modify the directory where network captures, port scans, and analysis results are stored.
#### Save Network Config
Save the current network configuration settings for future use.
#### Load Network Config
Load and apply a previously saved network configuration.
## Network Recon
Listen stealthily for network traffic and perform initial discovery and enumeration.
#### Wireshark Listening
Short Network capture using Wireshark and perform analysis with tshark.
#### Discovery Scan
Perform a discovery scan using multiple tools to identify hosts and services on the network. Attempt to categorize any found endpoints.

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
