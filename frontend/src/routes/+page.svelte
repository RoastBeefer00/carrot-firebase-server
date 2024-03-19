<script>
    import Recipe from "$lib/components/recipe.svelte";
    import { recipes, filter, searchRecipesByName, searchRecipesByIngredient } from "$lib/stores.js";

    let search;

    async function handleKeyUp(event) {
        if (event.key === 'Enter') {
            if ($filter === 'name') {
                await searchRecipesByName(search);
            } else {
                await searchRecipesByIngredient(search)
            }
            search = '';
        }
    };
</script>

<div class="mt-4 sm:mt-1">
    <div>
        <input
            class="flex-initial w-full shadow rounded appearance-none border-8 border-lavender p-2 outline-none"
            bind:value={search}
            on:keyup={handleKeyUp}
            type="search"
            placeholder="Press 'Enter' to search by {$filter}..."
        />
    </div>
    <div>
        {#if $recipes != []}
            {#each $recipes as recipe, index}
                <div class="p-2">
                    <Recipe {recipe} {index} />
                </div>
            {/each}
        {/if}
    </div>
</div>
