<script>
    import { groceries, getGroceries } from "$lib/stores.js";
    import { onMount } from "svelte";

    onMount(async () => {
        await getGroceries();
    });
</script>

<div class="h-full w-full rounded bg-crust p-2">
    <div class="flex flex-col rounded bg-lavender h-full">
        <div class="bg-crust text-text p-4">
            <h1 class="text-3xl font-bold">Groceries</h1>
        </div>
        <div class="p-4 overflow-y-auto">
            {#each $groceries as grocery}
                <div class="flex p-2">
                    <input type="checkbox" />
                    <span class="ml-2">
                        {grocery.quantity}
                        {grocery.measurement}
                        {grocery.item}
                    </span>
                </div>
            {/each}
        </div>
    </div>
</div>
