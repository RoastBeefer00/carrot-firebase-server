<script>
    import Recipe from "$lib/components/recipe.svelte";

    const apiUrl = (path) => `${import.meta.env.VITE_API_URL || ""}${path}`;

    const getRandomRecipe = async () => {
        const url = apiUrl("/recipes/random");
        console.log(url);
        const res = await fetch(url);
        if (!res.ok) {
            throw "Error while fetching data from ${url} (${res.status} ${res.statusText}).`;";
        }
        const recipe = await res.json();
        console.log(recipe);
        return recipe;
    };
</script>

{#await getRandomRecipe()}
    loading...
{:then recipe}
    <Recipe recipe={recipe} />
{:catch err}
    {err}
{/await}
