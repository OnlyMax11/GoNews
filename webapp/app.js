document.addEventListener('DOMContentLoaded', () => {
    const container = document.getElementById('news-container');
    const refreshBtn = document.getElementById('refresh-btn');
    
    // Загружаем новости при старте
    loadNews(10);
    
    // Обработчик кнопки обновления
    refreshBtn.addEventListener('click', () => {
        container.innerHTML = '<div class="loading">Loading...</div>';
        loadNews(10);
    });
    
    async function loadNews(count) {
        try {
            const response = await fetch(`/news/${count}`);
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const posts = await response.json();
            renderPosts(posts);
        } catch (error) {
            console.error('Error:', error);
            container.innerHTML = `
                <div class="error">
                    <p>Error loading news. Please try again later.</p>
                    <button onclick="location.reload()">Retry</button>
                </div>
            `;
        }
    }
    
    function renderPosts(posts) {
        if (!posts || posts.length === 0) {
            container.innerHTML = '<p class="no-news">No news available</p>';
            return;
        }
        
        container.innerHTML = posts.map(post => `
            <div class="news-card">
                ${post.imageUrl ? `
                <img src="${post.imageUrl}" alt="${post.title}" class="news-image" onerror="this.style.display='none'">
                ` : ''}
                <div class="news-content">
                    <h3 class="news-title">${escapeHtml(post.title)}</h3>
                    <div class="news-meta">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                            <path d="M3 6h18M7 12h10M5 18h14"></path>
                        </svg>
                        ${formatDate(post.pubTime)}
                    </div>
                    <p class="news-description">${truncateText(extractTextFromHTML(post.content), 150)}</p>
                    <a href="${escapeHtml(post.link)}" class="read-more" target="_blank">Read full article</a>
                </div>
            </div>
        `).join('');
    }
    
    // Вспомогательные функции
    function escapeHtml(str) {
        if (!str) return '';
        return str.replace(/[&<>'"]/g, 
            tag => ({
                '&': '&amp;',
                '<': '&lt;',
                '>': '&gt;',
                "'": '&#39;',
                '"': '&quot;'
            }[tag] || tag));
    }
    
    function formatDate(timestamp) {
        return new Date(timestamp * 1000).toLocaleString('ru-RU', {
            day: 'numeric',
            month: 'numeric',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    }
    
    function truncateText(text, maxLength) {
        if (!text) return '';
        return text.length > maxLength ? 
            text.substring(0, maxLength) + '...' : text;
    }
    
    function extractTextFromHTML(html) {
        if (!html) return '';
        const doc = new DOMParser().parseFromString(html, 'text/html');
        return doc.body.textContent || "";
    }
});