# 🚀 RELEASE NOTICE: NEXA Ultimate v3.1
**"The Genesis Update"** | *February 5, 2026*

---

We are proud to announce the immediate availability of **NEXA Ultimate v3.1**, a monumental leap forward in local decentralized networking. This release introduces the **Unified Genesis Core**, reducing resource footprint by 60% while delivering a completely redesigned, professional-grade user experience.

## ✨ Highlights (أبرز التحديثات)

### 1. 🏗️ Unified Core Architecture (النواة الموحدة)
Gone are the days of managing multiple conflicting service windows.
*   **Single Binary:** All 7 subsystems (Server, Gateway, Storage, Chat, Admin, DNS, Dashboard) now live inside a single, optimized executable: `nexa.exe`.
*   **Zero-Latency:** Internal services now communicate over shared memory channels and Go-routines.
*   **Smart Orchestration:** Automatic failure recovery and parallel service booting.

### 2. 💎 Storage 2.0 (السحابة الاحترافية)
A total rewrite of the file management system.
*   **UI Overhaul:** Stunning Glassmorphism interface with dark mode.
*   **Pro Features:**
    *   **Vault 🔐:** Secure, PIN-protected folder for sensitive data.
    *   **Auto-Backup 🔄:** Background daemon automatically safeguards your `incoming` files every 5 minutes.
    *   **Smart Sharing 🔗:** Instant QR Code generation for mobile transfer & short-links.
    *   **Live Search 🔍:** Real-time filtering by name and file type (Images, Videos, Docs).

### 3. 🛡️ Networking & Connectivity
*   **Dashboard Proxy Fix:** Solved all CORS and connectivity issues when accessing tools via the main dashboard.
*   **Unified Port Map:** Standardized ports configuration across the entire stack.

---

## 📦 What's Included? (محتويات الحزمة)

| Component | Status | Version |
| :--- | :---: | :---: |
| **Nexa Core Engine** | ✅ Stable | 3.1.0 |
| **Storage Service** | ✅ Stable | 2.5 (Pro) |
| **Quantum Chat** | ✅ Stable | 1.2 |
| **Matrix Dashboard** | ✅ Stable | 3.0 |
| **Admin Panel** | ⚠️ Beta | 0.9 |

---

## 🛠️ How to Upgrade (طريقة التحديث)

Since the architecture has changed significantly, a **Clean Build** is required.

1.  **Stop** any running Nexa instances.
2.  **Run** the build script:
    ```cmd
    BUILD.bat
    ```
3.  **Launch** the new unified core:
    ```cmd
    bin\start-all.bat
    ```
4.  **Access** the new Command center:
    👉 `http://localhost:7000`

---

## 📝 Developer Note
> "This update transforms Nexa from a collection of tools into a cohesive Operating System. We focused heavily on the 'Feel' of the software—making it not just functional, but beautiful and enjoyable to use."

**Happy Networking!** 🌐
*The Nexa Development Team*
