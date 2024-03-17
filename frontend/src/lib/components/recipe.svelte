<script>
    export let recipe;
</script>

<div class="bg-lavender p-4 rounded">
    <div>
        <h1 class="text-mantle text-3xl font-bold">{recipe.name}</h1>
    </div>
    <div>
        <span class="text-mantle">{recipe.time}</span>
    </div>
    <div class="bg-surface0 m-2 p-2 rounded">
        <ul>
            {#each recipe.ingredients as ingredient}
                <li class="text-text">{ingredient}</li>
            {/each}
        </ul>
    </div>
    <div class="bg-surface0 m-2 p-2 rounded">
        {#each recipe.steps as step}
            <div>
                <span class="text-text">{step}</span>
            </div>
        {/each}
    </div>
</div>
