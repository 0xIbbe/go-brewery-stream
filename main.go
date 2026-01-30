package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Brewery struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	BreweryType   string   `json:"brewery_type"`
	Address1      string   `json:"address_1"`
	City          string   `json:"city"`
	StateProvince string   `json:"state_province"`
	PostalCode    string   `json:"postal_code"`
	Country       string   `json:"country"`
	Longitude     *float64 `json:"longitude"`
	Latitude      *float64 `json:"latitude"`
	Phone         string   `json:"phone"`
	WebsiteURL    string   `json:"website_url"`
}

type StreamBrewery struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	TypeColor   string `json:"typeColor"`
	Address     string `json:"address"`
	City        string `json:"city"`
	State       string `json:"state"`
	Country     string `json:"country"`
	Phone       string `json:"phone"`
	Website     string `json:"website"`
	MapURL      string `json:"mapUrl"`
	HasLocation bool   `json:"hasLocation"`
}

var tmpl = template.Must(template.New("index").Parse(htmlTemplate))

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	r.Get("/", handleHome)
	r.Get("/stream", handleStream)
	r.Get("/health", handleHealth)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🍺 Brewery Stream running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, nil)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok","service":"brewery-stream"}`))
}

func handleStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	sendBrewery(w, flusher)

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			sendBrewery(w, flusher)
		}
	}
}

func sendBrewery(w http.ResponseWriter, flusher http.Flusher) {
	brewery, err := fetchRandomBrewery()
	if err != nil {
		log.Printf("Error fetching brewery: %v", err)
		return
	}

	data, err := json.Marshal(brewery)
	if err != nil {
		log.Printf("Error marshaling brewery: %v", err)
		return
	}

	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

func fetchRandomBrewery() (*StreamBrewery, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get("https://api.openbrewerydb.org/v1/breweries/random?size=1")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var breweries []Brewery
	if err := json.NewDecoder(resp.Body).Decode(&breweries); err != nil {
		return nil, err
	}

	if len(breweries) == 0 {
		return nil, fmt.Errorf("no brewery returned")
	}

	b := breweries[0]
	return transformBrewery(b), nil
}

func transformBrewery(b Brewery) *StreamBrewery {
	address := b.Address1
	if address == "" {
		address = "Address not available"
	}

	phone := b.Phone
	if phone == "" {
		phone = "Not available"
	} else {
		phone = formatPhone(phone)
	}

	website := b.WebsiteURL
	if website == "" {
		website = ""
	}

	state := b.StateProvince
	if state == "" {
		state = "N/A"
	}

	city := b.City
	if city == "" {
		city = "Unknown"
	}

	country := b.Country
	if country == "" {
		country = "Unknown"
	}

	hasLocation := b.Latitude != nil && b.Longitude != nil
	mapURL := buildMapURL(b.Name, b.Address1, city, state, country)

	return &StreamBrewery{
		Name:        b.Name,
		Type:        formatBreweryType(b.BreweryType),
		TypeColor:   getTypeColor(b.BreweryType),
		Address:     address,
		City:        city,
		State:       state,
		Country:     country,
		Phone:       phone,
		Website:     website,
		MapURL:      mapURL,
		HasLocation: hasLocation,
	}
}

func buildMapURL(name, address, city, state, country string) string {
	var parts []string

	if name != "" {
		parts = append(parts, name)
	}
	if address != "" {
		parts = append(parts, address)
	}
	if city != "" {
		parts = append(parts, city)
	}
	if state != "" && state != "N/A" {
		parts = append(parts, state)
	}
	if country != "" && country != "Unknown" {
		parts = append(parts, country)
	}

	if len(parts) == 0 {
		return ""
	}

	query := url.QueryEscape(strings.Join(parts, ", "))
	return fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", query)
}

func formatBreweryType(t string) string {
	types := map[string]string{
		"micro":      "Micro Brewery",
		"nano":       "Nano Brewery",
		"regional":   "Regional Brewery",
		"brewpub":    "Brewpub",
		"large":      "Large Brewery",
		"planning":   "Planning",
		"bar":        "Bar",
		"contract":   "Contract Brewing",
		"proprietor": "Proprietor",
		"closed":     "Closed",
	}
	if name, ok := types[t]; ok {
		return name
	}
	return t
}

func getTypeColor(t string) string {
	colors := map[string]string{
		"micro":      "bg-amber-500",
		"nano":       "bg-yellow-500",
		"regional":   "bg-orange-500",
		"brewpub":    "bg-green-500",
		"large":      "bg-blue-500",
		"planning":   "bg-purple-500",
		"bar":        "bg-pink-500",
		"contract":   "bg-indigo-500",
		"proprietor": "bg-teal-500",
		"closed":     "bg-gray-500",
	}
	if color, ok := colors[t]; ok {
		return color
	}
	return "bg-gray-500"
}

func formatPhone(phone string) string {
	if len(phone) == 10 {
		return fmt.Sprintf("(%s) %s-%s", phone[:3], phone[3:6], phone[6:])
	}
	return phone
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Brewery Stream | Real-time Discovery</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
    <style>
        body { font-family: 'Inter', sans-serif; }
        .beer-gradient { background: linear-gradient(135deg, #f59e0b 0%, #d97706 50%, #92400e 100%); }
        .glass-effect { 
            background: rgba(255, 255, 255, 0.95);
            backdrop-filter: blur(10px);
        }
        @keyframes pour {
            0% { transform: translateY(-20px); opacity: 0; }
            100% { transform: translateY(0); opacity: 1; }
        }
        @keyframes bubble {
            0%, 100% { transform: translateY(0) scale(1); }
            50% { transform: translateY(-10px) scale(1.1); }
        }
        .pour-in { animation: pour 0.6s ease-out; }
        .bubble { animation: bubble 2s ease-in-out infinite; }
        .pulse-ring {
            animation: pulse-ring 2s cubic-bezier(0.455, 0.03, 0.515, 0.955) infinite;
        }
        @keyframes pulse-ring {
            0% { transform: scale(0.8); opacity: 1; }
            100% { transform: scale(2); opacity: 0; }
        }
    </style>
</head>
<body class="bg-gradient-to-br from-amber-50 via-orange-50 to-yellow-50 min-h-screen">
    <div class="container mx-auto px-4 py-8 max-w-4xl">
        <!-- Header -->
        <header class="text-center mb-10">
            <div class="inline-flex items-center justify-center mb-4">
                <span class="text-6xl bubble">🍺</span>
            </div>
            <h1 class="text-4xl md:text-5xl font-bold text-transparent bg-clip-text beer-gradient bg-gradient-to-r from-amber-600 to-orange-700 mb-3">
                Brewery Stream
            </h1>
            <p class="text-amber-700 text-lg max-w-xl mx-auto">
                Discover craft breweries from around the world in real-time
            </p>
            <div class="mt-4 inline-flex items-center gap-2 px-4 py-2 bg-amber-100 rounded-full">
                <div class="relative">
                    <div class="w-2 h-2 bg-green-500 rounded-full"></div>
                    <div class="absolute inset-0 w-2 h-2 bg-green-400 rounded-full pulse-ring"></div>
                </div>
                <span class="text-amber-800 text-sm font-medium">
                    Streaming from 8,000+ breweries worldwide
                </span>
            </div>
        </header>

        <!-- Controls -->
        <div class="flex justify-center gap-4 mb-8">
            <button id="startBtn" onclick="startStream()" 
                class="px-6 py-3 bg-gradient-to-r from-amber-500 to-orange-500 text-white rounded-xl font-semibold shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 transition-all duration-200 flex items-center gap-2">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"></path>
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                </svg>
                Start Discovery
            </button>
            <button id="stopBtn" onclick="stopStream()" disabled
                class="px-6 py-3 bg-gray-200 text-gray-400 rounded-xl font-semibold cursor-not-allowed transition-all duration-200 flex items-center gap-2">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z"></path>
                </svg>
                Stop
            </button>
        </div>

        <!-- Stats -->
        <div class="grid grid-cols-3 gap-4 mb-8">
            <div class="glass-effect rounded-xl p-4 text-center shadow-lg border border-amber-100">
                <div class="text-3xl font-bold text-amber-600" id="discoveredCount">0</div>
                <div class="text-amber-700 text-sm">Discovered</div>
            </div>
            <div class="glass-effect rounded-xl p-4 text-center shadow-lg border border-amber-100">
                <div class="text-3xl font-bold text-orange-600" id="countriesCount">0</div>
                <div class="text-amber-700 text-sm">Countries</div>
            </div>
            <div class="glass-effect rounded-xl p-4 text-center shadow-lg border border-amber-100">
                <div class="text-3xl font-bold text-yellow-600" id="typesCount">0</div>
                <div class="text-amber-700 text-sm">Types</div>
            </div>
        </div>

        <!-- Current Brewery Card -->
        <div id="breweryCard" class="hidden glass-effect rounded-2xl shadow-2xl overflow-hidden border border-amber-100 mb-8 pour-in">
            <div class="p-6 md:p-8">
                <!-- Type Badge -->
                <div class="flex items-center justify-between mb-4">
                    <span id="breweryType" class="px-4 py-1.5 bg-amber-500 text-white rounded-full text-sm font-semibold">
                        Type
                    </span>
                    <span id="breweryCountry" class="text-amber-600 font-medium flex items-center gap-1">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"></path>
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"></path>
                        </svg>
                        <span></span>
                    </span>
                </div>

                <!-- Name -->
                <h2 id="breweryName" class="text-2xl md:text-3xl font-bold text-gray-800 mb-4">
                    Brewery Name
                </h2>

                <!-- Details Grid -->
                <div class="grid md:grid-cols-2 gap-4 mb-6">
                    <!-- Address -->
                    <div class="flex items-start gap-3 p-4 bg-amber-50 rounded-xl">
                        <div class="p-2 bg-amber-100 rounded-lg">
                            <svg class="w-5 h-5 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"></path>
                            </svg>
                        </div>
                        <div>
                            <div class="text-xs text-amber-600 font-medium uppercase tracking-wider mb-1">Address</div>
                            <div id="breweryAddress" class="text-gray-700">Address</div>
                            <div id="breweryCity" class="text-gray-600 text-sm">City, State</div>
                        </div>
                    </div>

                    <!-- Phone -->
                    <div class="flex items-start gap-3 p-4 bg-orange-50 rounded-xl">
                        <div class="p-2 bg-orange-100 rounded-lg">
                            <svg class="w-5 h-5 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"></path>
                            </svg>
                        </div>
                        <div>
                            <div class="text-xs text-orange-600 font-medium uppercase tracking-wider mb-1">Phone</div>
                            <div id="breweryPhone" class="text-gray-700">Phone number</div>
                        </div>
                    </div>
                </div>

                <!-- Action Buttons -->
                <div class="flex flex-wrap gap-3">
                    <a id="mapLink" href="#" target="_blank" 
                        class="inline-flex items-center gap-2 px-5 py-2.5 bg-gradient-to-r from-blue-500 to-blue-600 text-white rounded-xl font-medium hover:shadow-lg transition-all duration-200">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7"></path>
                        </svg>
                        View on Map
                    </a>
                    <a id="websiteLink" href="#" target="_blank" 
                        class="hidden inline-flex items-center gap-2 px-5 py-2.5 bg-gradient-to-r from-amber-500 to-orange-500 text-white rounded-xl font-medium hover:shadow-lg transition-all duration-200">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"></path>
                        </svg>
                        Visit Website
                    </a>
                </div>
            </div>
        </div>

        <!-- Placeholder -->
        <div id="placeholder" class="glass-effect rounded-2xl shadow-xl p-12 text-center border border-amber-100">
            <div class="text-6xl mb-4">🍻</div>
            <h3 class="text-xl font-semibold text-gray-700 mb-2">Ready to Explore?</h3>
            <p class="text-gray-500">Click "Start Discovery" to begin streaming breweries from around the world!</p>
        </div>

        <!-- History -->
        <div id="historySection" class="hidden mt-8">
            <h3 class="text-xl font-semibold text-amber-800 mb-4 flex items-center gap-2">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                </svg>
                Recently Discovered
            </h3>
            <div id="historyList" class="space-y-3"></div>
        </div>

        <!-- Footer -->
        <footer class="mt-12 text-center text-amber-600 text-sm">
            <p>Powered by <a href="https://www.openbrewerydb.org/" target="_blank" class="underline hover:text-amber-800">OpenBreweryDB</a></p>
            <p class="mt-1">Real-time streaming with Server-Sent Events</p>
        </footer>
    </div>

    <script>
        let eventSource = null;
        let discoveredCount = 0;
        let countries = new Set();
        let types = new Set();
        let history = [];

        function startStream() {
            if (eventSource) return;

            eventSource = new EventSource('/stream');
            
            document.getElementById('startBtn').disabled = true;
            document.getElementById('startBtn').className = 'px-6 py-3 bg-gray-200 text-gray-400 rounded-xl font-semibold cursor-not-allowed transition-all duration-200 flex items-center gap-2';
            
            document.getElementById('stopBtn').disabled = false;
            document.getElementById('stopBtn').className = 'px-6 py-3 bg-gradient-to-r from-red-500 to-red-600 text-white rounded-xl font-semibold shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 transition-all duration-200 flex items-center gap-2';
            
            document.getElementById('placeholder').classList.add('hidden');

            eventSource.onmessage = function(event) {
                const brewery = JSON.parse(event.data);
                displayBrewery(brewery);
                updateStats(brewery);
                addToHistory(brewery);
            };

            eventSource.onerror = function() {
                console.log('SSE connection error, reconnecting...');
            };
        }

        function stopStream() {
            if (eventSource) {
                eventSource.close();
                eventSource = null;
            }

            document.getElementById('stopBtn').disabled = true;
            document.getElementById('stopBtn').className = 'px-6 py-3 bg-gray-200 text-gray-400 rounded-xl font-semibold cursor-not-allowed transition-all duration-200 flex items-center gap-2';
            
            document.getElementById('startBtn').disabled = false;
            document.getElementById('startBtn').className = 'px-6 py-3 bg-gradient-to-r from-amber-500 to-orange-500 text-white rounded-xl font-semibold shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 transition-all duration-200 flex items-center gap-2';
        }

        function displayBrewery(brewery) {
            const card = document.getElementById('breweryCard');
            card.classList.remove('hidden');
            card.classList.remove('pour-in');
            void card.offsetWidth;
            card.classList.add('pour-in');

            document.getElementById('breweryName').textContent = brewery.name;
            
            const typeEl = document.getElementById('breweryType');
            typeEl.textContent = brewery.type;
            typeEl.className = 'px-4 py-1.5 ' + brewery.typeColor + ' text-white rounded-full text-sm font-semibold';
            
            document.getElementById('breweryCountry').querySelector('span').textContent = brewery.country;
            document.getElementById('breweryAddress').textContent = brewery.address;
            document.getElementById('breweryCity').textContent = brewery.city + ', ' + brewery.state;
            document.getElementById('breweryPhone').textContent = brewery.phone;
            
            const mapLink = document.getElementById('mapLink');
            if (brewery.mapUrl) {
                mapLink.href = brewery.mapUrl;
                mapLink.classList.remove('hidden');
            } else {
                mapLink.classList.add('hidden');
            }

            const websiteLink = document.getElementById('websiteLink');
            if (brewery.website) {
                websiteLink.href = brewery.website;
                websiteLink.classList.remove('hidden');
            } else {
                websiteLink.classList.add('hidden');
            }
        }

        function updateStats(brewery) {
            discoveredCount++;
            countries.add(brewery.country);
            types.add(brewery.type);

            document.getElementById('discoveredCount').textContent = discoveredCount;
            document.getElementById('countriesCount').textContent = countries.size;
            document.getElementById('typesCount').textContent = types.size;
        }

        function addToHistory(brewery) {
            history.unshift(brewery);
            if (history.length > 5) history.pop();

            const section = document.getElementById('historySection');
            section.classList.remove('hidden');

            const list = document.getElementById('historyList');
            list.innerHTML = history.map(b => 
                '<div class="glass-effect rounded-xl p-4 border border-amber-100 flex items-center justify-between">' +
                    '<div>' +
                        '<div class="font-semibold text-gray-800">' + b.name + '</div>' +
                        '<div class="text-sm text-gray-500">' + b.city + ', ' + b.country + '</div>' +
                    '</div>' +
                    '<span class="px-3 py-1 ' + b.typeColor + ' text-white rounded-full text-xs font-medium">' + b.type + '</span>' +
                '</div>'
            ).join('');
        }
    </script>
</body>
</html>`
