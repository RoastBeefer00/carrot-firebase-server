<script>
    // import { base } from "$app/paths";
    import "../app.css";
</script>

<nav class="fixed top-0 z-30 w-full bg-base border-b border-rosewater">
    <div class="px-3 py-3 lg:px-5 lg:pl-3">
        <div class="flex items-center justify-between">
            <div
                class="flex items-center justify-start sm:justify-center w-full"
            >
                <button
                    class="inline-flex items-center p-2 mt-2 ml-3 text-sm text-gray-500 rounded-lg sm:hidden hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-gray-200 dark:text-gray-400 dark:hover:bg-gray-700 dark:focus:ring-gray-600"
                >
                    <span class="sr-only"> "Open sidebar" </span>
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
    class="fixed top-0 left-0 z-20 w-64 pt-20 h-screen transition-transform -translate-x-full sm:translate-x-0"
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
                <a
                    href="#"
                    class="flex items-center p-2 mb-2 text-crust rounded-lg bg-red"
                >
                    <span class="ml-4"> Remove All </span>
                </a>
            </li>
        </ul>
    </div>
</aside>

<div class="p-4 sm:ml-64 mt-16 sm:mt-28">
    <slot />
</div>
