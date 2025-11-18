// API Base URL
const API_BASE = 'http://localhost:3000';

// State Management
let authToken = localStorage.getItem('authToken') || null;
let currentUser = JSON.parse(localStorage.getItem('currentUser')) || null;
let currentPostId = null;

// Initialize app
document.addEventListener('DOMContentLoaded', () => {
    if (authToken) {
        showAuthenticatedUI();
        loadPosts();
    } else {
        showSection('home');
        loadPosts();
    }
    
    // Close modals when clicking outside
    document.getElementById('viewPostSection').addEventListener('click', (e) => {
        if (e.target.id === 'viewPostSection') closePostModal();
    });
    document.getElementById('editPostSection').addEventListener('click', (e) => {
        if (e.target.id === 'editPostSection') closeEditModal();
    });
});

// Section Navigation
function showSection(section) {
    // Hide all sections
    document.querySelectorAll('main section').forEach(s => s.classList.add('hidden'));
    
    // Show selected section
    const sections = {
        'login': 'loginSection',
        'register': 'registerSection',
        'home': 'homeSection',
        'create': 'createSection',
        'profile': 'profileSection'
    };
    
    const sectionId = sections[section];
    if (sectionId) {
        document.getElementById(sectionId).classList.remove('hidden');
    }
    
    // Load data for specific sections
    if (section === 'home') {
        loadPosts();
    } else if (section === 'profile' && authToken) {
        loadProfile();
        loadMyPosts();
    }
    
    // Scroll to top
    window.scrollTo({ top: 0, behavior: 'smooth' });
}

// Show/Hide UI elements based on auth state
function showAuthenticatedUI() {
    document.getElementById('mainNav').classList.remove('hidden');
    document.getElementById('authButtons').classList.add('hidden');
}

function showUnauthenticatedUI() {
    document.getElementById('mainNav').classList.add('hidden');
    document.getElementById('authButtons').classList.remove('hidden');
}

// Toast Notification
function showToast(message, duration = 3000) {
    const toast = document.getElementById('toast');
    const toastMessage = document.getElementById('toastMessage');
    
    toastMessage.textContent = message;
    toast.classList.remove('hidden');
    toast.classList.add('toast-show');
    
    setTimeout(() => {
        toast.classList.remove('toast-show');
        toast.classList.add('toast-hide');
        setTimeout(() => {
            toast.classList.add('hidden');
            toast.classList.remove('toast-hide');
        }, 300);
    }, duration);
}

// Authentication Handlers
async function handleLogin(event) {
    event.preventDefault();
    
    const email = document.getElementById('loginEmail').value;
    const password = document.getElementById('loginPassword').value;
    
    try {
        const response = await fetch(`${API_BASE}/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });
        
        if (response.ok) {
            const data = await response.json();
            authToken = data.token;
            currentUser = { email };
            
            localStorage.setItem('authToken', authToken);
            localStorage.setItem('currentUser', JSON.stringify(currentUser));
            
            showToast('Login successful!');
            showAuthenticatedUI();
            showSection('home');
        } else {
            const error = await response.text();
            showToast('Login failed: ' + error);
        }
    } catch (error) {
        showToast('Error: ' + error.message);
    }
}

async function handleRegister(event) {
    event.preventDefault();
    
    const name = document.getElementById('registerName').value;
    const email = document.getElementById('registerEmail').value;
    const password = document.getElementById('registerPassword').value;
    
    try {
        const response = await fetch(`${API_BASE}/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, email, password })
        });
        
        if (response.ok) {
            showToast('Registration successful! Please login.');
            showSection('login');
        } else {
            const error = await response.text();
            showToast('Registration failed: ' + error);
        }
    } catch (error) {
        showToast('Error: ' + error.message);
    }
}

function logout() {
    authToken = null;
    currentUser = null;
    localStorage.removeItem('authToken');
    localStorage.removeItem('currentUser');
    
    showUnauthenticatedUI();
    showToast('Logged out successfully');
    showSection('home');
}

// Posts Management
async function loadPosts() {
    const container = document.getElementById('postsContainer');
    container.innerHTML = '<div class="col-span-full flex justify-center"><div class="spinner"></div></div>';
    
    try {
        const response = await fetch(`${API_BASE}/posts`);
        
        if (response.ok) {
            const posts = await response.json();
            displayPosts(posts || []);
        } else {
            container.innerHTML = '<p class="text-gray-500 text-center col-span-full">Failed to load posts</p>';
        }
    } catch (error) {
        container.innerHTML = '<p class="text-gray-500 text-center col-span-full">Error loading posts</p>';
    }
}

function displayPosts(posts) {
    const container = document.getElementById('postsContainer');
    
    if (posts.length === 0) {
        container.innerHTML = `
            <div class="col-span-full text-center py-20">
                <p class="text-gray-600 text-xl mb-4">No posts yet. Be the first to create one!</p>
                ${authToken ? `<button onclick="showSection('create')" class="px-6 py-3 bg-black text-white rounded-lg hover:bg-gray-800 font-bold transition-all hover:scale-105">Create Post</button>` : ''}
            </div>
        `;
        return;
    }
    
    container.innerHTML = posts.map(post => `
        <div class="post-card bg-white border-4 border-black rounded-2xl shadow-xl overflow-hidden hover:shadow-2xl transition-all cursor-pointer group" onclick="viewPost(${post.id})">
            ${post.thumbnail ? `
                <div class="relative overflow-hidden h-48">
                    <img src="${post.thumbnail}" alt="${post.title}" 
                         class="w-full h-full object-cover group-hover:scale-110 transition-transform duration-500">
                </div>
            ` : `
                <div class="w-full h-48 bg-gradient-to-br from-gray-100 to-gray-200 flex items-center justify-center border-b-4 border-black">
                    <svg class="w-16 h-16 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z"></path>
                    </svg>
                </div>
            `}
            <div class="p-6">
                <div class="flex items-center justify-between mb-3">
                    <span class="px-3 py-1 text-xs font-bold rounded-full ${
                        post.status === 'published' 
                            ? 'bg-black text-white' 
                            : 'bg-gray-300 text-gray-800'
                    }">
                        ${post.status.toUpperCase()}
                    </span>
                    <span class="text-sm text-gray-500">${formatDate(post.created_at)}</span>
                </div>
                <h3 class="text-xl font-bold text-black mb-2 line-clamp-2 group-hover:text-gray-700 transition-colors">${escapeHtml(post.title)}</h3>
                <p class="text-black line-clamp-3 mb-4 leading-relaxed">${escapeHtml(post.content)}</p>
                <div class="flex items-center text-black group-hover:translate-x-2 transition-transform">
                    <span class="font-bold">Read more</span>
                    <svg class="w-5 h-5 ml-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 8l4 4m0 0l-4 4m4-4H3"></path>
                    </svg>
                </div>
            </div>
        </div>
    `).join('');
}

async function handleCreatePost(event) {
    event.preventDefault();
    
    if (!authToken) {
        showToast('Please login to create a post');
        showSection('login');
        return;
    }
    
    const title = document.getElementById('postTitle').value;
    const content = document.getElementById('postContent').value;
    const thumbnail = document.getElementById('postThumbnail').value || null;
    const status = document.getElementById('postStatus').value;
    
    try {
        const response = await fetch(`${API_BASE}/posts`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify({ title, content, thumbnail, status })
        });
        
        if (response.ok) {
            showToast('Post created successfully!');
            document.getElementById('postTitle').value = '';
            document.getElementById('postContent').value = '';
            document.getElementById('postThumbnail').value = '';
            document.getElementById('postStatus').value = 'draft';
            showSection('home');
        } else {
            const error = await response.text();
            showToast('Failed to create post: ' + error);
        }
    } catch (error) {
        showToast('Error: ' + error.message);
    }
}

async function viewPost(postId) {
    currentPostId = postId;
    const modal = document.getElementById('viewPostSection');
    const content = document.getElementById('postDetailContent');
    
    content.innerHTML = '<div class="flex justify-center py-10"><div class="spinner"></div></div>';
    modal.classList.remove('hidden');
    
    try {
        const response = await fetch(`${API_BASE}/posts/${postId}`);
        
        if (response.ok) {
            const post = await response.json();
            content.innerHTML = `
                ${post.thumbnail ? `
                    <div class="relative overflow-hidden rounded-xl mb-6 border-4 border-black">
                        <img src="${post.thumbnail}" alt="${post.title}" 
                             class="w-full h-80 object-cover">
                    </div>
                ` : ''}
                <div class="flex items-center justify-between mb-6">
                    <span class="px-4 py-1.5 text-sm font-bold rounded-full ${
                        post.status === 'published' 
                            ? 'bg-black text-white' 
                            : 'bg-gray-300 text-gray-800'
                    }">
                        ${post.status.toUpperCase()}
                    </span>
                    <span class="text-gray-600">${formatDate(post.created_at)}</span>
                </div>
                <h1 class="text-4xl font-black text-black mb-6 leading-tight">${escapeHtml(post.title)}</h1>
                <div class="prose max-w-none text-black whitespace-pre-wrap leading-relaxed text-lg">${escapeHtml(post.content)}</div>
                <div class="mt-8 pt-6 border-t-2 border-black">
                    <p class="text-sm text-gray-600">
                        Last updated: ${formatDate(post.updated_at)}
                    </p>
                </div>
            `;
        } else {
            content.innerHTML = '<p class="text-red-600 text-center py-10 font-bold">Failed to load post</p>';
        }
    } catch (error) {
        content.innerHTML = '<p class="text-red-600 text-center py-10 font-bold">Error loading post</p>';
    }
}

function closePostModal() {
    document.getElementById('viewPostSection').classList.add('hidden');
    currentPostId = null;
}

async function loadMyPosts() {
    if (!authToken) return;
    
    const container = document.getElementById('myPostsContainer');
    container.innerHTML = '<div class="flex justify-center py-4"><div class="spinner"></div></div>';
    
    try {
        const response = await fetch(`${API_BASE}/my-posts`, {
            headers: { 'Authorization': `Bearer ${authToken}` }
        });
        
        if (response.ok) {
            const posts = await response.json();
            displayMyPosts(posts || []);
        } else {
            container.innerHTML = '<p class="text-gray-500 text-center py-4">Failed to load your posts</p>';
        }
    } catch (error) {
        container.innerHTML = '<p class="text-gray-500 text-center py-4">Error loading posts</p>';
    }
}

function displayMyPosts(posts) {
    const container = document.getElementById('myPostsContainer');
    
    if (posts.length === 0) {
        container.innerHTML = `
            <div class="text-center py-12">
                <p class="text-gray-600 mb-4">You haven't created any posts yet.</p>
                <button onclick="showSection('create')" class="px-6 py-3 bg-black text-white rounded-lg hover:bg-gray-800 font-bold transition-all hover:scale-105">
                    Create Your First Post
                </button>
            </div>
        `;
        return;
    }
    
    container.innerHTML = `
        <div class="space-y-4">
            ${posts.map(post => `
                <div class="bg-white border-2 border-black rounded-xl p-5 hover:shadow-xl transition-all">
                    <div class="flex justify-between items-start mb-3">
                        <div class="flex-1">
                            <h4 class="text-lg font-bold text-black mb-2">${escapeHtml(post.title)}</h4>
                            <p class="text-black text-sm line-clamp-2">${escapeHtml(post.content)}</p>
                        </div>
                        <span class="px-3 py-1 text-xs font-bold rounded-full ml-4 ${
                            post.status === 'published' 
                                ? 'bg-black text-white' 
                                : 'bg-gray-300 text-gray-800'
                        }">
                            ${post.status.toUpperCase()}
                        </span>
                    </div>
                    <div class="flex justify-between items-center mt-4 pt-4 border-t-2 border-gray-200">
                        <span class="text-xs text-gray-500">${formatDate(post.created_at)}</span>
                        <div class="flex gap-2">
                            <button onclick="viewPost(${post.id})" 
                                    class="px-4 py-1.5 text-sm text-black hover:bg-gray-100 rounded-lg border-2 border-black transition-colors font-semibold">
                                View
                            </button>
                            <button onclick="editPost(${post.id})" 
                                    class="px-4 py-1.5 text-sm text-white bg-black hover:bg-gray-800 rounded-lg transition-colors font-bold">
                                Edit
                            </button>
                            <button onclick="deletePost(${post.id})" 
                                    class="px-4 py-1.5 text-sm text-red-600 hover:bg-red-50 rounded-lg border-2 border-red-600 transition-colors font-semibold">
                                Delete
                            </button>
                        </div>
                    </div>
                </div>
            `).join('')}
        </div>
    `;
}

async function editPost(postId) {
    try {
        const response = await fetch(`${API_BASE}/posts/${postId}`);
        
        if (response.ok) {
            const post = await response.json();
            document.getElementById('editPostId').value = post.id;
            document.getElementById('editPostTitle').value = post.title;
            document.getElementById('editPostContent').value = post.content;
            document.getElementById('editPostThumbnail').value = post.thumbnail || '';
            document.getElementById('editPostStatus').value = post.status;
            
            document.getElementById('editPostSection').classList.remove('hidden');
        } else {
            showToast('Failed to load post for editing');
        }
    } catch (error) {
        showToast('Error: ' + error.message);
    }
}

function closeEditModal() {
    document.getElementById('editPostSection').classList.add('hidden');
}

async function handleUpdatePost(event) {
    event.preventDefault();
    
    const postId = document.getElementById('editPostId').value;
    const title = document.getElementById('editPostTitle').value;
    const content = document.getElementById('editPostContent').value;
    const thumbnail = document.getElementById('editPostThumbnail').value || null;
    const status = document.getElementById('editPostStatus').value;
    
    try {
        const response = await fetch(`${API_BASE}/posts/${postId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify({ title, content, thumbnail, status })
        });
        
        if (response.ok) {
            showToast('Post updated successfully!');
            closeEditModal();
            loadMyPosts();
            loadPosts();
        } else {
            const error = await response.text();
            showToast('Failed to update post: ' + error);
        }
    } catch (error) {
        showToast('Error: ' + error.message);
    }
}

async function deletePost(postId) {
    if (!confirm('Are you sure you want to delete this post?')) {
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE}/posts/${postId}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${authToken}`
            }
        });
        
        if (response.ok) {
            showToast('Post deleted successfully!');
            loadMyPosts();
            loadPosts();
        } else {
            const error = await response.text();
            showToast('Failed to delete post: ' + error);
        }
    } catch (error) {
        showToast('Error: ' + error.message);
    }
}

// Profile
function loadProfile() {
    const profileInfo = document.getElementById('profileInfo');
    
    if (currentUser) {
        profileInfo.innerHTML = `
            <div class="space-y-6">
                <div class="flex items-center space-x-5">
                    <div class="w-24 h-24 bg-black rounded-full flex items-center justify-center text-white text-3xl font-black shadow-lg border-4 border-black">
                        ${currentUser.email.charAt(0).toUpperCase()}
                    </div>
                    <div>
                        <h3 class="text-2xl font-bold text-black mb-1">${currentUser.email}</h3>
                        <p class="text-gray-600 font-medium">Content Creator</p>
                    </div>
                </div>
                <div class="border-t-2 border-black pt-5 space-y-2">
                    <p class="text-gray-700"><span class="font-bold text-black">Email:</span> ${currentUser.email}</p>
                    <p class="text-gray-700"><span class="font-bold text-black">Member since:</span> ${new Date().toLocaleDateString()}</p>
                </div>
            </div>
        `;
    }
}

// Utility Functions
function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { 
        year: 'numeric', 
        month: 'short', 
        day: 'numeric' 
    });
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
