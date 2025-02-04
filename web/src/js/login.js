document.addEventListener('DOMContentLoaded', () => {
    const form = document.querySelector('.login-form');
    
    if (form) {
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const username = form.querySelector('#username').value;
            const password = form.querySelector('#password').value;
            const submitButton = form.querySelector('button[type="submit"]');
            
            // Disable the submit button
            submitButton.disabled = true;
            
            try {
                const response = await fetch('/admin/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/x-www-form-urlencoded',
                    },
                    body: new URLSearchParams({
                        username,
                        password
                    })
                });
                
                if (response.redirected) {
                    window.location.href = response.url;
                } else {
                    const data = await response.text();
                    document.documentElement.innerHTML = data;
                }
            } catch (error) {
                console.error('Login error:', error);
                // Show error message
                const errorDiv = document.createElement('div');
                errorDiv.className = 'login-form__error';
                errorDiv.textContent = 'An error occurred. Please try again.';
                form.insertBefore(errorDiv, submitButton);
            } finally {
                submitButton.disabled = false;
            }
        });
    }
});
