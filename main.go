package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Item struct {
	ID          int     `json:"id"`
	Name        string  `json:name`
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}
type Inventory struct {
	items  []Item
	nextID int
	mu     sync.Mutex
}

func NewInventory() *Inventory {
	return &Inventory{
		items:  []Item{},
		nextID: 1,
	}
}
func (inv *Inventory) CreateItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid Json", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(item.Name) == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if item.Quantity < 0 {
		http.Error(w, "quantity cannot be negative", http.StatusBadRequest)
		return
	}
	if item.Price < 0 {
		http.Error(w, "price cannot be negative", http.StatusBadRequest)
		return
	}
	inv.mu.Lock()
	item.ID = inv.nextID
	inv.nextID++
	inv.items = append(inv.items, item)
	inv.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func (inv *Inventory) GetItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	inv.mu.Lock()
	defer inv.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inv.items)
}
func (inv *Inventory) GetItem(w http.ResponseWriter, r *http.Request, id int) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	inv.mu.Lock()
	defer inv.mu.Unlock()

	for _, item := range inv.items {
		if item.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	http.Error(w, "Item not found", http.StatusNotFound)
}
func (inv *Inventory) UpdateItem(w http.ResponseWriter, r *http.Request, id int) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	var updatedItem Item
	if err := json.NewDecoder(r.Body).Decode(&updatedItem); err != nil {
		http.Error(w, "Invalid Json", http.StatusBadRequest)
		return
	}
	inv.mu.Lock()
	defer inv.mu.Unlock()
	for i, item := range inv.items {
		if item.ID == id {
			if strings.TrimSpace(updatedItem.Name) != "" {
				inv.items[i].Name = updatedItem.Name
			}
			if strings.TrimSpace(updatedItem.Description) != "" {
				inv.items[i].Description = updatedItem.Description
			}
			if updatedItem.Quantity >= 0 {
				inv.items[i].Quantity = updatedItem.Quantity
			}
			if updatedItem.Price >= 0 {
				inv.items[i].Price = updatedItem.Price
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(inv.items[i])
			return
		}
	}
	http.Error(w, "Item not found", http.StatusNotFound)
}

func (inv *Inventory) DeleteItem(w http.ResponseWriter, r *http.Request, id int) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	inv.mu.Lock()
	defer inv.mu.Unlock()
	for i, item := range inv.items {
		if item.ID == id {
			inv.items = append(inv.items[:i], inv.items[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.Error(w, "Item not found", http.StatusNotFound)
}

func (inv *Inventory) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) == 1 && segments[0] == "items" {
		switch method {
		case http.MethodGet:
			inv.GetItems(w, r)
		case http.MethodPost:
			inv.CreateItem(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	if len(segments) == 2 && segments[0] == "items" {
		id, err := strconv.Atoi(segments[1])
		if err != nil {
			http.Error(w, "Invalid Item ID", http.StatusBadRequest)
			return
		}
		switch method {
		case http.MethodGet:
			inv.GetItem(w, r, id)
		case http.MethodPut:
			inv.UpdateItem(w, r, id)
		case http.MethodDelete:
			inv.DeleteItem(w, r, id)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	http.Error(w, "Not found", http.StatusNotFound)
}

func main() {
	inventory := NewInventory()
	mux := http.NewServeMux()
	mux.Handle("/itrems", inventory)
	mux.Handle("/items", inventory)
	port := ":4455"
	fmt.Println("Server running at port %s...\n", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
