# 🌌 Nexa OS - Ultimate Local Cloud System (v3.0)

> **The Next-Generation Local Network Operating System.**  
> *Transforming standard LAN connectivity into a fully-featured, intelligent cloud infrastructure.*

---

## 📋 Table of Contents
1. [Project Overview](#-project-overview)
2. [System Architecture](#-system-architecture)
3. [Key Features](#-key-features)
4. [Microservices Breakdown](#-microservices-breakdown)
5. [User Interface (UI/UX)](#-user-interface-uiyx)
6. [Networking & Security](#-networking--security)
7. [Installation & Usage](#-installation--usage)
8. [Credits & Developers](#-credits--developers)

---

## 🔭 Project Overview

**Nexa OS** is not just a file-sharing tool; it is a **complete local cloud ecosystem** built from scratch using **Go (Golang)**. It replaces traditional OS networking limitations (like SMB/Windows Sharing) with a web-based, platform-independent architecture.

The system acts as a "mini-internet" inside your local network, featuring its own **DNS Authority**, **Central Gateway**, **Cloud Storage**, and **Social Platform (Chat)**.

### 🚀 What's New in v3.0 Ultimate?
- **Transition from CLI to GUI:** No more staring at terminal screens; everything is now web-based.
- **Mobile First:** Full support for Smartphones, Tablets, and Laptops via Hotspot/Wi-Fi.
- **Intelligent Networking:** Auto-detection of Host IP and Firewall Bypassing.
- **Unified Dashboard:** A single glassmorphism interface to control the entire system.

---

## 🏗 System Architecture

The project follows a **Microservices Architecture**, where each component runs independently but communicates via a central nervous system (The Gateway).

```mermaid
graph TD
    User[User Device (Mobile/PC)] --> Gateway[Central Gateway (:8000)]
    Gateway --> Dashboard[Unified Dashboard (:7000)]
    Gateway --> Cloud[Nexa Cloud Storage (:8081)]
    Gateway --> Chat[Live Chat System (:8082)]
    Gateway --> Admin[Admin Controller (:8080)]
    Gateway --> DNS[DNS Authority (:53)]
    Gateway --> Core[Ledger Core (:9000)]
```

---

## 🌟 Key Features

### 1. 📂 Nexa Cloud Storage (The Local Drive)
*   **Web-Based Interface:** Upload/Download files from any browser (Chrome, Safari, Edge).
*   **Drag & Drop:** Supports uploading huge files simply by dragging them.
*   **Media Streaming:** Watch videos or listen to music directly from the server without downloading.
*   **Cross-Device:** Transfer files from Phone to PC (and vice versa) instantly via Wi-Fi.

### 2. 💬 Nexa Live Chat (The Social Hub)
*   **Real-Time Messaging:** Instant communication system for everyone connected to the network.
*   **Zero-Setup:** Auto-login (Guest) system; just type your name and chat.
*   **Persistent History:** Keeps the last 1000 messages for new joiners to see context.

### 3. 🖥️ The Ultimate Dashboard (Control Center)
*   **Glassmorphism Design:** A stunning, modern UI inspired by Windows 11 and MacOS.
*   **Live Monitoring:** Shows active services, storage usage, and network status.
*   **One-Click Navigation:** Jump between Admin, Files, and Chat instantly.

### 4. 🧠 Custom DNS Authority
*   **Why IP Addresses?** Instead of `192.168.1.5`, use `.nexa` domains.
*   **Resolution:** A fully functional DNS server running on Port 53 (UDP/TCP).
*   **Service Discovery:** Automatically registers new services (e.g., `chat.nexa`).

---

## 🧩 Microservices Breakdown

| Service | Port | Description | Tech Stack |
|:---|:---:|:---|:---|
| **Gateway** | `8000` | The entry point Main Router. Handles traffic & proxying. | Go (Chi Router) |
| **Dashboard** | `7000` | The User Interface frontend. | HTML5/JS/CSS3 |
| **Admin Panel**| `8080` | System control, logs, and user management. | Go + Bcrypt |
| **Web Storage**| `8081` | File Server logic (Upload/Download). | Go + Multipart/Form |
| **Chat Svc** | `8082` | JSON API for messaging and real-time sync. | Go + Mutex Locks |
| **Core Node** | `9000` | Blockchain-based ledger for critical data. | Go + TCP Sockets |
| **DNS Server** | `53` | Domain Name System implementation. | Go + UDP |

---

## 🎨 User Interface (UI/UX)
The interface was designed with **"Wow Factor"** in mind.

- **Visual Style:** Deep Space Dark Mode with Neon Accents (Cyan/Purple).
- **Glass Effect:** Heavy use of `backdrop-filter: blur()` for that "Frosted Glass" look.
- **Responsiveness:** Fluid grid layouts that adapt from 4K Monitors down to 5-inch Phone screens.
- **Animations:** Smooth CSS transitions for hover effects, modal openings, and message arrival.

---

## 🛡️ Networking & Security

### 🧱 Firewall Bypass System
We developed custom **Automation Scripts** to handle Windows Firewall restrictions:
- `fix-firewall.bat`: Automatically opens Ports 7000, 8000, 8080, 8081, 8082, 9000.
- **Hotspot Support:** Configured to work on `Public` network profiles (common in Mobile Hotspots).

### 🔒 Security Layers
- **Admin Auth:** Bcrypt hashing for password protection on the Admin Panel.
- **CORS Policies:** Configured middlewares to allow secure Cross-Origin Resource Sharing.
- **Input Sanitization:** Protection against basic injection attacks in the Chat and File modules.

---

## 📥 Installation & Usage

### Prerequisites
- Windows 10/11 (for the Host)
- Any device with a Browser (Client)
- Go 1.22+ (for building only)

### 🚀 Quick Start
1. **Build the System:**
   ```powershell
   .\bin\build-linux.bat  # For Linux
   # OR use the manual Go build commands for Windows
   ```

2. **Fix Network Permissions:**
   Right-click `fix-firewall.bat` and **Run as Administrator**. (One time only)

3. **Launch Nexa OS:**
   Double click `bin\start-all.bat`.
   > *All terminal windows will open with professional ASCII banners.*

4. **Connect:**
   - **On Host:** Go to `http://localhost:7000`
   - **On Mobile:** Go to `http://[YOUR-PC-IP]:7000`

---

## © Credits & Developers

This massive project was brought to life by:

### 👑 MultiX0
> **Core Developer & Founder**
> *Responsible for the original idea, the Blockchain Core, and the initial CLI Protocol design.*

### 🛠️ Islam Ibrahim
> **System Architect & Full-Stack Developer**
> *Responsible for the v3.0 Ultimate Upgrade: The Web Ecosystem, Unified Dashboard, Gateway Architecture, Chat System, UI/UX Design, and Networking Solutions.*

---

*Made with ❤️ and a lot of caffeine.*
*Nexa OS - v3.0 Ultimate Build*