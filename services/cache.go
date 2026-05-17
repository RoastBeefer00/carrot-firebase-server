package services

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type RecipeCache struct {
	mu      sync.RWMutex
	recipes []Recipe
	ready   chan struct{}
	once    sync.Once
}

func NewRecipeCache() *RecipeCache {
	return &RecipeCache{ready: make(chan struct{})}
}

func (c *RecipeCache) Start(ctx context.Context, client *firestore.Client) {
	go c.run(ctx, client)
}

func (c *RecipeCache) run(ctx context.Context, client *firestore.Client) {
	for {
		if err := c.listen(ctx, client); err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("recipe cache listener error: %v; reconnecting in 5s", err)
			select {
			case <-time.After(5 * time.Second):
			case <-ctx.Done():
				return
			}
		}
	}
}

func (c *RecipeCache) listen(ctx context.Context, client *firestore.Client) error {
	it := client.Collection("recipes").Snapshots(ctx)
	defer it.Stop()
	for {
		snap, err := it.Next()
		if err != nil {
			return err
		}
		recipes := make([]Recipe, 0, snap.Size)
		docs := snap.Documents
		for {
			doc, err := docs.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Printf("recipe cache doc iter error: %v", err)
				break
			}
			var r Recipe
			if err := doc.DataTo(&r); err != nil {
				log.Printf("recipe cache decode error for %s: %v", doc.Ref.ID, err)
				continue
			}
			recipes = append(recipes, r)
		}
		c.mu.Lock()
		c.recipes = recipes
		c.mu.Unlock()
		c.once.Do(func() { close(c.ready) })
		log.Printf("recipe cache updated: %d recipes", len(recipes))
	}
}

func (c *RecipeCache) Ready() <-chan struct{} {
	return c.ready
}

func (c *RecipeCache) WaitReady(timeout time.Duration) bool {
	select {
	case <-c.ready:
		return true
	case <-time.After(timeout):
		return false
	}
}

func (c *RecipeCache) All() []Recipe {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Recipe, len(c.recipes))
	copy(out, c.recipes)
	return out
}

func (c *RecipeCache) SearchByName(q string) []Recipe {
	q = strings.ToLower(q)
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []Recipe
	for _, r := range c.recipes {
		if strings.Contains(strings.ToLower(r.Name), q) {
			out = append(out, r)
		}
	}
	return out
}

func (c *RecipeCache) SearchByIngredient(q string) []Recipe {
	q = strings.ToLower(q)
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []Recipe
	for _, r := range c.recipes {
		for _, ing := range r.Ingredients {
			if strings.Contains(strings.ToLower(ing), q) {
				out = append(out, r)
				break
			}
		}
	}
	return out
}
