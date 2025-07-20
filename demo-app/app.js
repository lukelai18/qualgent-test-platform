// 全局变量
let currentUser = null;
let cart = [];
let currentStep = 1;
const totalSteps = 5;

// 页面导航函数
function showPage(pageId) {
    // 隐藏所有页面
    document.querySelectorAll('.page').forEach(page => {
        page.classList.remove('active');
    });
    
    // 显示指定页面
    document.getElementById(pageId).classList.add('active');
}

// 登录相关函数
function showLogin() {
    showPage('login-page');
}

function showSignup() {
    showPage('signup-page');
}

// 登录表单处理
document.getElementById('login-form').addEventListener('submit', function(e) {
    e.preventDefault();
    
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    
    // 清除之前的错误
    clearErrors();
    
    // 验证表单
    if (!username) {
        showError('username-error');
        return;
    }
    
    if (!password) {
        showError('password-error');
        return;
    }
    
    // 模拟登录验证
    if (username === 'testuser@example.com' && password === 'password123') {
        currentUser = {
            name: 'testuser',
            email: 'testuser@example.com'
        };
        showDashboard();
    } else if (username === 'invalid@example.com' && password === 'wrongpassword') {
        showErrorMessage('Invalid credentials');
    } else {
        showErrorMessage('用户名或密码错误');
    }
});

// 注册表单处理
document.getElementById('signup-form').addEventListener('submit', function(e) {
    e.preventDefault();
    
    const email = document.getElementById('signup-email').value;
    const password = document.getElementById('signup-password').value;
    const confirmPassword = document.getElementById('confirm-password').value;
    
    if (password !== confirmPassword) {
        alert('密码不匹配');
        return;
    }
    
    // 模拟注册成功
    showMessage('注册成功！验证邮件已发送');
    showLogin();
});

// 引导流程函数
function nextStep() {
    if (currentStep < totalSteps) {
        // 验证当前步骤
        if (currentStep === 2) {
            const firstName = document.getElementById('first-name').value;
            const email = document.getElementById('profile-email').value;
            
            if (!firstName) {
                showError('first-name-error');
                return;
            }
            
            if (!email) {
                showError('email-error');
                return;
            }
        }
        
        // 隐藏当前步骤
        document.querySelector(`[data-step="${currentStep}"]`).classList.remove('active');
        
        // 显示下一步
        currentStep++;
        document.querySelector(`[data-step="${currentStep}"]`).classList.add('active');
        
        // 特殊处理
        if (currentStep === 4) {
            // 显示验证码
            document.querySelector('[data-testid="verification-code"]').style.display = 'block';
        }
    }
}

function verifyEmail() {
    const code = document.querySelector('[data-testid="verification-input"]').value;
    
    if (code === '123456') {
        nextStep();
    } else {
        alert('验证码错误');
    }
}

function skipOnboarding() {
    if (confirm('确定要跳过引导设置吗？')) {
        showDashboard();
    }
}

function goToDashboard() {
    showDashboard();
}

// 主页面函数
function showDashboard() {
    showPage('dashboard-page');
    updateUserInfo();
}

function showProducts() {
    showPage('products-page');
    updateCartCount();
}

function showCart() {
    showPage('cart-page');
    updateCartDisplay();
}

function showCheckout() {
    showPage('checkout-page');
}

function showOrders() {
    showPage('orders-page');
}

function logout() {
    currentUser = null;
    cart = [];
    showLogin();
}

// 购物车函数
function addToCart(productId) {
    const products = {
        1: { name: '笔记本电脑', price: 8999 },
        2: { name: '智能手机', price: 4999 },
        3: { name: '无线耳机', price: 1299 }
    };
    
    const product = products[productId];
    cart.push(product);
    updateCartCount();
    
    // 显示添加成功消息
    alert(`${product.name} 已添加到购物车`);
}

function updateCartCount() {
    const cartCountElements = document.querySelectorAll('[data-testid="cart-count"]');
    cartCountElements.forEach(element => {
        element.textContent = cart.length;
    });
}

function updateCartDisplay() {
    const cartItems = document.getElementById('cart-items');
    const cartTotal = document.getElementById('cart-total');
    
    cartItems.innerHTML = '';
    let total = 0;
    
    cart.forEach((item, index) => {
        const cartItem = document.createElement('div');
        cartItem.className = 'cart-item';
        cartItem.innerHTML = `
            <span>${item.name}</span>
            <span>¥${item.price}</span>
            <button onclick="removeFromCart(${index})">删除</button>
        `;
        cartItems.appendChild(cartItem);
        total += item.price;
    });
    
    cartTotal.textContent = total;
}

function removeFromCart(index) {
    cart.splice(index, 1);
    updateCartCount();
    updateCartDisplay();
}

// 搜索函数
function searchProducts() {
    const searchTerm = document.querySelector('[data-testid="search-input"]').value;
    alert(`搜索: ${searchTerm}`);
}

// 结账函数
function saveShipping() {
    const name = document.getElementById('shipping-name').value;
    const address = document.getElementById('shipping-address').value;
    const city = document.getElementById('shipping-city').value;
    const zip = document.getElementById('shipping-zip').value;
    
    if (!name || !address || !city || !zip) {
        alert('请填写完整的配送信息');
        return;
    }
    
    alert('配送信息已保存');
}

function processPayment() {
    const cardNumber = document.getElementById('card-number').value;
    const cardExpiry = document.getElementById('card-expiry').value;
    const cardCvv = document.getElementById('card-cvv').value;
    
    if (!cardNumber || !cardExpiry || !cardCvv) {
        alert('请填写完整的支付信息');
        return;
    }
    
    // 模拟支付处理
    if (cardNumber === '4000000000000002') {
        // 支付失败
        document.querySelector('[data-testid="payment-error"]').style.display = 'block';
        document.querySelector('[data-testid="payment-error"]').textContent = 'Payment failed';
    } else {
        // 支付成功
        showPage('order-success-page');
        cart = []; // 清空购物车
        updateCartCount();
    }
}

// 用户信息更新
function updateUserInfo() {
    if (currentUser) {
        document.getElementById('user-name').textContent = currentUser.name;
        document.getElementById('user-email').textContent = currentUser.email;
    }
}

// 错误处理函数
function clearErrors() {
    document.querySelectorAll('.error').forEach(error => {
        error.style.display = 'none';
    });
    document.querySelectorAll('.error-message').forEach(error => {
        error.style.display = 'none';
    });
}

function showError(errorId) {
    document.querySelector(`[data-testid="${errorId}"]`).style.display = 'block';
}

function showErrorMessage(message) {
    const errorElement = document.querySelector('[data-testid="error-message"]');
    errorElement.textContent = message;
    errorElement.style.display = 'block';
}

function showMessage(message) {
    alert(message);
}

// 初始化
document.addEventListener('DOMContentLoaded', function() {
    // 设置默认用户（用于演示）
    currentUser = {
        name: 'testuser',
        email: 'testuser@example.com'
    };
    
    // 显示登录页面
    showLogin();
    
    // 更新购物车计数
    updateCartCount();
}); 