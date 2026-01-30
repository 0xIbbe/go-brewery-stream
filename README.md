# 🍺 Brewery Stream

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![SSE](https://img.shields.io/badge/SSE-Real--time-FF6B6B?style=for-the-badge&logo=lightning&logoColor=white)
![Tailwind CSS](https://img.shields.io/badge/Tailwind-3.0-38B2AC?style=for-the-badge&logo=tailwind-css&logoColor=white)
![API](https://img.shields.io/badge/API-OpenBreweryDB-F59E0B?style=for-the-badge)
![Render](https://img.shields.io/badge/Render-Deployed-46E3B7?style=for-the-badge&logo=render&logoColor=white)

Real-time craft brewery discovery stream using **Go**, **Server-Sent Events (SSE)**, and **OpenBreweryDB API**. Explore **8,000+ breweries** worldwide with live streaming and beautiful Tailwind CSS UI!

## ✨ Features

- 🌊 **Real-time SSE streaming** - Continuous brewery discovery
- 🍻 **8,000+ breweries** - From micro to large, worldwide coverage
- 🗺️ **Google Maps integration** - View brewery locations
- 📊 **Live statistics** - Track discoveries, countries, types
- 📜 **Discovery history** - Recent breweries at a glance
- 🎨 **Beautiful UI** - Glass morphism with Tailwind CSS
- 📱 **Fully responsive** - Works on all devices
- ⚡ **Chi router** - Fast and lightweight

## 🚀 Quick Start

Clone the repository:

```bash
git clone https://github.com/smart-developer1791/go-brewery-stream
cd go-brewery-stream
```

Initialize dependencies and run:

```bash
go mod tidy
go run .
```

Open http://localhost:8080 in your browser.

## 🏗️ Tech Stack

| Technology | Purpose |
|------------|---------|
| **Go 1.21+** | Backend server |
| **Chi Router** | HTTP routing & middleware |
| **SSE** | Real-time streaming |
| **OpenBreweryDB** | Brewery data API |
| **Tailwind CSS** | Styling |

## 📡 API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /` | Main UI page |
| `GET /stream` | SSE brewery stream |
| `GET /health` | Health check |

## 🍺 Brewery Types

| Type | Description |
|------|-------------|
| 🏭 **Micro** | < 15,000 barrels/year |
| 🧪 **Nano** | < 200 barrels/year |
| 🏢 **Regional** | 15,000-6M barrels/year |
| 🍽️ **Brewpub** | Restaurant-brewery |
| 🏗️ **Large** | > 6M barrels/year |
| 📋 **Contract** | Contract brewing |

## 📁 Project Structure

```text
go-brewery-stream/
├── main.go          # Application entry point
├── go.mod           # Go module file
├── render.yaml      # Render deployment config
├── .gitignore       # Git ignore rules
└── README.md        # Documentation
```

## 🔧 Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |

## 🌐 How SSE Works

```text
Client                    Server
  |                          |
  |------- GET /stream ----->|
  |                          |
  |<---- data: {...} --------|  (brewery 1)
  |<---- data: {...} --------|  (brewery 2)
  |<---- data: {...} --------|  (brewery 3)
  |         ...              |
```

## 📊 Data Source

This project uses the free [OpenBreweryDB API](https://www.openbrewerydb.org/):

- 🌍 8,000+ breweries worldwide
- 🆓 Free and open source
- 🔄 Regularly updated
- 📍 Location coordinates included

---

## Deploy in 10 seconds

[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy)
