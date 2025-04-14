// Configure marked.js for consistent parsing
marked.setOptions({
    gfm: true, // Enable GitHub Flavored Markdown
    breaks: true, // Support line breaks
    pedantic: false,
    smartLists: true,
    smartypants: false
});

document.addEventListener('DOMContentLoaded', () => {
    fetch('content.json')
        .then(response => response.json())
        .then(data => {
            const sidebar = document.getElementById('sidebar');
            const navLinks = document.getElementById('nav-links');
            const content = document.getElementById('content');
            const sidebarToggle = document.getElementById('sidebar-toggle');
            const sidebarOpen = document.getElementById('sidebar-open');

            // Generate Sidebar
            Object.keys(data).forEach(key => {
                if (key !== 'title') {
                    const link = document.createElement('a');
                    link.href = `#${key}`;
                    link.className = 'block text-blue-600 hover:text-blue-800 transition-colors duration-200 py-1';
                    link.textContent = data[key].title;
                    link.dataset.section = key;
                    navLinks.appendChild(link);
                }
            });

            // Generate Content
            content.innerHTML = `
                <h1 class="text-4xl font-extrabold text-gray-900 mb-6 animate-fade-in">${data.title}</h1>
                ${Object.keys(data).filter(k => k !== 'title').map(key => {
                const section = data[key];
                if (key === 'examples') {
                    return `
                            <section id="${key}" class="mb-12 animate-fade-in">
                                <h2 class="text-3xl font-semibold text-gray-800 mb-6">${section.title}</h2>
                                ${section.items.map(item => `
                                    <div class="mb-8">
                                        <h3 class="text-xl font-medium text-gray-700 mb-3">${item.title}</h3>
                                        <p class="text-gray-600 mb-4">${item.description}</p>
                                        <div class="relative">
                                            <pre><code class="language-go">${item.code}</code></pre>
                                            <button class="copy-button" onclick="copyCode(this)">Copy</button>
                                        </div>
                                    </div>
                                `).join('')}
                            </section>
                        `;
                } else if (key === 'use_cases') {
                    return `
                            <section id="${key}" class="mb-12 animate-fade-in">
                                <h2 class="text-3xl font-semibold text-gray-800 mb-6">${section.title}</h2>
                                <ul class="list-disc pl-6 text-gray-600">
                                    ${section.items.map(item => `
                                        <li class="mb-3"><span class="font-medium text-gray-700">${item.title}:</span> ${item.content}</li>
                                    `).join('')}
                                </ul>
                            </section>
                        `;
                } else if (key === 'faq') {
                    return `
                            <section id="${key}" class="mb-12 animate-fade-in">
                                <h2 class="text-3xl font-semibold text-gray-800 mb-6">${section.title}</h2>
                                <div class="space-y-4">
                                    ${section.content.map(item => `
                                        <div>
                                            <h3 class="text-lg font-medium text-gray-700 mb-2">${item.question}</h3>
                                            <p class="text-gray-600">${item.answer}</p>
                                        </div>
                                    `).join('')}
                                </div>
                            </section>
                        `;
                } else {
                    return `
                            <section id="${key}" class="mb-12 animate-fade-in">
                                <h2 class="text-3xl font-semibold text-gray-800 mb-6">${section.title}</h2>
                                ${Array.isArray(section.content) ? `
                                    <ul class="list-disc pl-6 text-gray-600">
                                        ${section.content.map(item => `<li class="mb-3">${marked.parseInline(item)}</li>`).join('')}
                                    </ul>
                                ` : `
                                    <div class="markdown-body">${marked.parse(section.content)}</div>
                                `}
                            </section>
                        `;
                }
            }).join('')}
            `;

            // Initialize Prism.js
            Prism.highlightAll();

            // Sidebar Toggle for Mobile
            sidebarOpen.addEventListener('click', () => {
                sidebar.classList.remove('-translate-x-full');
            });
            sidebarToggle.addEventListener('click', () => {
                sidebar.classList.add('-translate-x-full');
            });

            // Smooth Scrolling
            document.querySelectorAll('a[href^="#"]').forEach(anchor => {
                anchor.addEventListener('click', e => {
                    e.preventDefault();
                    const target = document.querySelector(anchor.getAttribute('href'));
                    target.scrollIntoView({ behavior: 'smooth' });
                    if (window.innerWidth < 768) {
                        sidebar.classList.add('-translate-x-full');
                    }
                });
            });

            // Active Section Highlighting
            const sections = document.querySelectorAll('section');
            const navItems = document.querySelectorAll('#nav-links a');
            window.addEventListener('scroll', () => {
                let current = '';
                sections.forEach(section => {
                    const sectionTop = section.offsetTop;
                    if (pageYOffset >= sectionTop - 60) {
                        current = section.getAttribute('id');
                    }
                });
                navItems.forEach(item => {
                    item.classList.remove('font-semibold', 'text-blue-800');
                    if (item.dataset.section === current) {
                        item.classList.add('font-semibold', 'text-blue-800');
                    }
                });
            });
        })
        .catch(error => console.error('Error loading content:', error));
});

// Copy Code Function
function copyCode(button) {
    const code = button.previousElementSibling.querySelector('code').textContent;
    navigator.clipboard.writeText(code).then(() => {
        button.textContent = 'Copied!';
        setTimeout(() => button.textContent = 'Copy', 2000);
    });
}