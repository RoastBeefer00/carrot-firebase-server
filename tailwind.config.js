/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./**/*.html", "./**/*.templ", "./**/*.go"],
  plugins: [require("@catppuccin/tailwindcss")],
}

