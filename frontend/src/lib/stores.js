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

export const deleteAllRecipes = async () => {
    recipes.set([]);
}

export function deleteRecipe(recipe) {
    recipes.update((recipes) => recipes.filter((r) => r !== recipe));
};

export const recipes = writable([]);
