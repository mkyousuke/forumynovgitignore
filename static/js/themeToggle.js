document.addEventListener('DOMContentLoaded', () => {
  const toggleBtn = document.getElementById('theme-toggle');
  if (!toggleBtn) return; // sécurité si le bouton n’existe pas

  const body = document.body;
  const savedTheme = localStorage.getItem('theme');

  if (savedTheme === 'dark') {
    body.classList.add('dark-mode');
    toggleBtn.textContent = '☀'; 
  }

  toggleBtn.addEventListener('click', () => {
    body.classList.toggle('dark-mode');
    if (body.classList.contains('dark-mode')) {
      toggleBtn.textContent = '☀';
      localStorage.setItem('theme', 'dark');
    } else {
      toggleBtn.textContent = '🌙';
      localStorage.setItem('theme', 'light');
    }
  });
});
