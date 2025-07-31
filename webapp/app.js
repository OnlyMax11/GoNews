document.addEventListener('DOMContentLoaded', () => {
    const newsContainer = document.getElementById('news-container');
    const countInput = document.getElementById('news-count');
    const refreshBtn = document.getElementById('refresh-btn');
    
    // Загрузка новостей при запуске
    loadNews(10);
    
    // Обработчик кнопки обновления
    refreshBtn.addEventListener('click', () => {
        const count = parseInt(countInput.value) || 10;
        loadNews(count);
    });
    
    function loadNews(count) {
        fetch(`/news/${count}`)
            .then(response => {
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .then(posts => renderPosts(posts))
            .catch(error => {
                console.error('Error:', error);
                newsContainer.innerHTML = `<p class="error">Error loading news. Please try again later.</p>`;
            });
    }
    
    function renderPosts(posts) {
        if (!posts || posts.length === 0) {
            newsContainer.innerHTML = `<p>No news available</p>`;
            return;
        }
        
        let html = '';
        posts.forEach(post => {
            const date = new Date(post.pubTime * 1000).toLocaleString();
            html += `
                <div class="post">
                    <h2>${escapeHTML(post.title)}</h2>
                    <div class="meta">${date}</div>
                    <div class="content">${truncateText(escapeHTML(post.content), 200)}</div>
                    <a href="${escapeHTML(post.link)}" target="_blank">Read full article</a>
                </div>
            `;
        });
        
        newsContainer.innerHTML = html;
    }
    
    function escapeHTML(str) {
        return str.replace(/[&<>"']/g, 
            tag => ({
                '&': '&amp;',
                '<': '&lt;',
                '>': '&gt;',
                '"': '&quot;',
                "'": '&#39;'
            }[tag] || tag));
    }
    
    function truncateText(text, maxLength) {
        if (text.length <= maxLength) return text;
        return text.substring(0, maxLength) + '...';
    }
});