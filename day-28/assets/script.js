// Day 27: Go Book Store — demo script
// Highlights the search query in product names after the page loads.
document.addEventListener('DOMContentLoaded', () => {
    const params = new URLSearchParams(window.location.search);
    const q = (params.get('q') || '').trim().toLowerCase();
    if (!q) return;

    document.querySelectorAll('tbody td:first-child').forEach(cell => {
        const text = cell.textContent;
        if (text.toLowerCase().includes(q)) {
            const re = new RegExp(`(${q})`, 'gi');
            cell.innerHTML = text.replace(re, '<mark>$1</mark>');
        }
    });
});
