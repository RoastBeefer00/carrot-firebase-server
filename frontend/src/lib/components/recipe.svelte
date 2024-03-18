<script>
    import { getRandomRecipe, deleteRecipe } from "$lib/stores.js";

    export let recipe;

    let open = false;

    function toggleOpen() {
        open = !open;
    }

    async function replaceRecipe() {
        recipe = await getRandomRecipe();
    }
</script>

<div class="bg-lavender p-4 rounded">
    <div class="flex mb-2">
        <h1 class="text-mantle text-3xl font-bold mr-2">{recipe.name}</h1>
        <span class="text-mantle mt-2">{recipe.time}</span>
    </div>
    <div class="flex justify-between">
        <button
            on:click={toggleOpen}
            class="text-crust text-center rounded bg-green p-2 border border-crust"
        >
            {#if !open}
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="32"
                    height="32"
                    viewBox="0 0 24 24"
                    ><path
                        fill="currentColor"
                        d="m12 18l-6-6l1.4-1.4l3.6 3.575V6h2v8.175l3.6-3.575L18 12Z"
                    /></svg
                >
            {:else}
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="32"
                    height="32"
                    viewBox="0 0 24 24"
                    ><path
                        fill="currentColor"
                        d="M11 18V9.825L7.4 13.4L6 12l6-6l6 6l-1.4 1.4L13 9.825V18Z"
                    /></svg
                >
            {/if}
        </button>
        <div class="flex gap-2">
            <button
                on:click={replaceRecipe}
                class="text-crust text-center rounded bg-sky p-2 border border-crust"
            >
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="32"
                    height="32"
                    viewBox="0 0 20 20"
                    ><g fill="currentColor"
                        ><path
                            d="M13.937 4.211a1 1 0 0 1-1.126 1.653A5 5 0 1 0 15 10a1 1 0 1 1 2 0a7 7 0 1 1-3.063-5.789"
                        /><path
                            d="M13.539 12.506a1 1 0 1 1-1.078-1.685l3.482-2.227a1 1 0 0 1 1.077 1.685z"
                        /><path
                            d="M18.903 12.41a1 1 0 0 1-1.826.815l-1.508-3.38a1 1 0 1 1 1.826-.815z"
                        /></g
                    ></svg
                >
            </button>
            <button
                on:click={() => deleteRecipe(recipe)}
                class="text-crust text-center rounded bg-red p-2 border border-crust"
            >
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="32"
                    height="32"
                    viewBox="0 0 24 24"
                    ><path
                        fill="currentColor"
                        d="M7 21q-.825 0-1.412-.587T5 19V6H4V4h5V3h6v1h5v2h-1v13q0 .825-.587 1.413T17 21zM17 6H7v13h10zM9 17h2V8H9zm4 0h2V8h-2zM7 6v13z"
                    /></svg
                >
            </button>
        </div>
    </div>
    {#if open}
        <div class="bg-surface2 m-2 p-2 rounded">
            <ul>
                {#each recipe.ingredients as ingredient}
                    <li class="text-text">{ingredient}</li>
                {/each}
            </ul>
        </div>
        <div class="bg-surface2 m-2 p-2 rounded">
            {#each recipe.steps as step}
                <div>
                    <span class="text-text">{step}</span>
                </div>
            {/each}
        </div>
    {/if}
</div>
