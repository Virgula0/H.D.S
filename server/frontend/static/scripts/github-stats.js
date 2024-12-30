document.addEventListener('DOMContentLoaded', function() {
    const repo = 'Virgula0/H.D.S';
    const apiUrl = `https://api.github.com/repos/${repo}`;

    fetch(apiUrl)
        .then(response => response.json())
        .then(data => {
            document.getElementById('stars').textContent = data.stargazers_count;
            document.getElementById('forks').textContent = data.forks_count;
            document.getElementById('issues').textContent = data.open_issues_count;

            const lastUpdate = new Date(data.updated_at);
            document.getElementById('last-update').textContent = lastUpdate.toLocaleDateString();
        })
        .catch(error => {
            console.error('Error fetching GitHub stats:', error);
            document.getElementById('github-stats').innerHTML = '<p class="text-danger">Error loading GitHub statistics</p>';
        });
});

