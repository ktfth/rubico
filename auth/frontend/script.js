const registerForm = document.getElementById('registerForm');
const messageDiv = document.getElementById('message');

registerForm.addEventListener('submit', (event) => {
    event.preventDefault(); // Impede o envio padrão do formulário

    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;

    // Requisição POST para o endpoint de registro
    fetch('/registerlogin', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ email, password })
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Erro na requisição');
            }
            return response.json();
        })
        .then(data => {
            messageDiv.textContent = data.message; // Exibe a mensagem de sucesso
        })
        .catch(error => {
            messageDiv.textContent = 'Erro ao registrar usuário'; // Exibe mensagem de erro genérica
            console.error(error); // Loga o erro no console
        });
});
