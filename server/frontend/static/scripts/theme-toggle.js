(function() {
    function createThemeToggle() {
        const button = document.createElement('button');
        button.className = 'btn btn-primary rounded-circle position-fixed';
        button.style.top = '20px';
        button.style.right = '20px';
        button.style.width = '40px';
        button.style.height = '40px';
        button.innerHTML = `
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="5"></circle>
                <line x1="12" y1="1" x2="12" y2="3"></line>
                <line x1="12" y1="21" x2="12" y2="23"></line>
                <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line>
                <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line>
                <line x1="1" y1="12" x2="3" y2="12"></line>
                <line x1="21" y1="12" x2="23" y2="12"></line>
                <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line>
                <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>
            </svg>
        `;
        document.body.appendChild(button);
        return button;
    }

    function toggleTheme() {
        document.body.classList.toggle('dark-mode');
        const theme = document.body.classList.contains('dark-mode') ? 'true' : 'false';
        localStorage.setItem('darkMode', theme);
    }

    function initTheme() {
        const currentTheme = localStorage.getItem('darkMode') || 'false';
        document.body.classList.toggle('dark-mode', currentTheme === 'true');
        const themeToggle = createThemeToggle();
        themeToggle.addEventListener('click', toggleTheme);
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initTheme);
    } else {
        initTheme();
    }
})();

