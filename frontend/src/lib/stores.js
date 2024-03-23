import { writable } from 'svelte/store';

const apiUrl = (path) => `${import.meta.env.VITE_API_URL || ""}${path}`;

export const getRandomRecipe = async () => {
    const url = apiUrl("/recipes/random");
    const res = await fetch(url);
    if (!res.ok) {
        throw "Error while fetching data from ${url} (${res.status} ${res.statusText}).`;";
    }
    const response = await res.json();
    return response;
};

export async function getRandomRecipes(amount) {
    const url = apiUrl("/recipes/random/" + amount);
    const res = await fetch(url);
    if (!res.ok) {
        throw "Error while fetching data from ${url} (${res.status} ${res.statusText}).`;";
    }
    const response = await res.json();

    for (let i = 0; i < response.length; i++) {
        recipes.update((recipes) => [...recipes, response[i]]);
    }
};

export async function replaceRecipe(index) {
    let newRecipe = await getRandomRecipe();
    console.log(newRecipe);
    recipes.update((recipes) => {
        recipes[index] = newRecipe;
        return recipes;
    });
}

export const deleteAllRecipes = async () => {
    recipes.set([]);
}

export function deleteRecipe(index) {
    recipes.update((recipes) => recipes.filter((_, i) => i != index));
};

export async function searchRecipesByName(name) {
    const url = apiUrl("/recipes/name/" + name);
    const res = await fetch(url);
    if (!res.ok) {
        throw "Error while fetching data from ${url} (${res.status} ${res.statusText}).`;";
    }
    const response = await res.json();

    for (let i = 0; i < response.length; i++) {
        recipes.update((recipes) => [...recipes, response[i]]);
    }

}

export async function searchRecipesByIngredient(ingredient) {
    const url = apiUrl("/recipes/ingredient/" + ingredient);
    const res = await fetch(url);
    if (!res.ok) {
        throw "Error while fetching data from ${url} (${res.status} ${res.statusText}).`;";
    }
    const response = await res.json();

    for (let i = 0; i < response.length; i++) {
        recipes.update((recipes) => [...recipes, response[i]]);
    }

}

export async function getGroceries() {
    let currentGroceries;
    recipes.subscribe((value) => (currentGroceries = value));
    const url = apiUrl("/groceries");
    const res = await fetch(url, {
        method: "POST",
        body: JSON.stringify(currentGroceries),
        headers: {
            "Content-type": "application/json; charset=UTF-8"
        }
    });

    if (!res.ok) {
        throw "Error while fetching data from ${url} (${res.status} ${res.statusText}).`;";
    }
    const response = await res.json();
    groceries.set(response);
}

export const recipes = writable([]);
export const groceries = writable([]);
export const filter = writable("name");
