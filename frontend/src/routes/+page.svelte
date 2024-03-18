<script>
    import Recipe from "$lib/components/recipe.svelte";
    import { recipes } from "$lib/stores.js";
</script>

<div>
{#if $recipes != []}
    {#each $recipes as recipe}
        <div class="p-2">
            <Recipe recipe={recipe} />
        </div>
    {/each}
{/if}
</div>
