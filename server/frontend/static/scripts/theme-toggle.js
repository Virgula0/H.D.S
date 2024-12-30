(function() {

    function toggleTheme() {
        const isDarkMode = document.body.classList.toggle('dark-mode');
        localStorage.setItem('darkMode', isDarkMode ? 'true' : 'false');
        updateThemeToggleIcon(isDarkMode);
    }

    function updateThemeToggleIcon(isDarkMode) {
        const icon = document.querySelector('#theme-toggle i');
        if (icon) {
            icon.className = isDarkMode ? 'fas fa-sun' : 'fas fa-moon';
        }
    }

    function initTheme() {
        const currentTheme = localStorage.getItem('darkMode') || 'false';
        const isDarkMode = currentTheme === 'true';
        document.body.classList.toggle('dark-mode', isDarkMode);
        updateThemeToggleIcon(isDarkMode);

        const themeToggle = document.getElementById('theme-toggle');
        if (themeToggle) {
            themeToggle.addEventListener('click', toggleTheme);
        }
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initTheme);
    } else {
        initTheme();
    }
})();

