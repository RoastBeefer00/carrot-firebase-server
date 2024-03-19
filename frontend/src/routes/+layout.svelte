<script>
    // import { base } from "$app/paths";
    import "../app.css";
    import { deleteAllRecipes, filter, getRandomRecipes } from "$lib/stores.js";

    let sidebarBase = "fixed top-0 left-0 z-20 w-64 pt-20 h-screen transition-transform sm:translate-x-0"
    let sidebarClass;
    let sidebarOpen = false;
    $: {
        sidebarClass = sidebarOpen ? sidebarBase + " translate-x-0" : sidebarBase + " -translate-x-full";
        console.log(sidebarClass);
    }

    let randomRecipes = 1;

    function toggleFilter() {
        if ($filter === "name") {
            filter.set("ingredient");
        } else {
            filter.set("name");
        };
    };
</script>

<nav class="fixed top-0 z-30 w-full bg-base border-b border-rosewater">
    <div class="px-3 py-3 lg:px-5 lg:pl-3">
        <div class="flex items-center justify-between">
            <div
                class="flex items-center justify-start sm:justify-center w-full"
            >
                <button
                    class="inline-flex items-center p-2 mt-2 ml-3 text-sm text-gray-500 rounded-lg sm:hidden hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-gray-200 dark:text-gray-400 dark:hover:bg-gray-700 dark:focus:ring-gray-600"
                    on:click={() => (sidebarOpen = !sidebarOpen)}
                >
                    <span class="sr-only"> "Open sidebar" </span>
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="32"
                        height="32"
                        viewBox="0 0 24 24"
                        ><path
                            fill="currentColor"
                            d="M3 18h18v-2H3zm0-5h18v-2H3zm0-7v2h18V6z"
                        /></svg
                    >
                </button>
                <h1
                    class="text-lavender lg:text-5xl text-4xl m-2 text-center font-extrabold"
                >
                    R&J Meals
                </h1>
            </div>
        </div>
    </div>
</nav>
<aside
    id="separator-sidebar"
    class={sidebarClass}
    aria-label="Sidebar"
>
    <div
        class="h-full px-3 sm:py-4 overflow-y-auto bg-base border-r border-r-rosewater"
    >
        <ul class="pt-4 sm:mt-4 space-y-2 font-medium">
            <li>
                <button
                    class="flex text-center p-2 text-crust rounded-lg bg-sky w-full"
                >
                    <span class="ml-4"> Grocery List </span>
                </button>
            </li>
            <li class="border-b border-b-rosewater">
                <button
                    on:click={deleteAllRecipes}
                    class="flex items-center p-2 mb-2 text-crust rounded-lg bg-red w-full"
                >
                    <span class="ml-4"> Remove All </span>
                </button>
            </li>
            <li class="border-b border-b-rosewater" >
                <label for="filters" class="block mb-2 text-sm font-medium text-gray-900 dark:text-white" >
                    Search by:
                </label>
                <select id="filters" on:change={toggleFilter} class="bg-gray-50 mb-2 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-lavender focus:border-lavender block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white" >
                    <option value="name">Name</option>
                    <option value="ingredients">Ingredient</option>
                </select>
            </li>
            <li>
                <label
                    for="exampleFormControlInputNumber"
                    class="pointer-events-none mb-0 max-w-[90%] truncate pt-[0.37rema] leading-[1.6] text-white transition-all duration-200 ease-out peer-focus:-translate-y-[0.9rem] peer-focus:scale-[0.8] peer-focus:text-primary peer-data-[te-input-state-active]:-translate-y-[0.9rem] peer-data-[te-input-state-active]:scale-[0.8] motion-reduce:transition-none dark:text-neutral-200 dark:peer-focus:text-primary"
                >
                    Add {randomRecipes} random recipes:
                </label>
                <div class="justify-between">
                    <input
                        type="number"
                        bind:value={randomRecipes}
                        min="0"
                        class="peer min-h-[auto] w-3/4 rounded border-0 bg-gray-700 px-3 py-[0.32rem] leading-[1.6] outline-none transition-all duration-200 ease-linear focus:placeholder:opacity-100 peer-focus:text-primary data-[te-input-state-active]:placeholder:opacity-100 motion-reduce:transition-none dark:text-white dark:placeholder:text-white dark:peer-focus:text-primary [&:not([data-te-input-placeholder-active])]:placeholder:opacity-0"
                    />
                    <button
                        class="p-1.5 rounded bg-gray-700 text-white float-right"
                        on:click={getRandomRecipes(randomRecipes)}
                    >
                        Add
                    </button>
                </div>
            </li>
        </ul>
    </div>
</aside>

<div class="p-4 sm:ml-64 mt-16 sm:mt-28">
    <slot />
</div>
