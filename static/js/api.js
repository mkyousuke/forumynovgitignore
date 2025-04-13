document.addEventListener("DOMContentLoaded", function() {
    // Récupérer tous les boutons "En savoir plus"
    const moreButtons = document.querySelectorAll(".more-btn");
    
    moreButtons.forEach(button => {
        button.addEventListener("click", function(e) {
            e.preventDefault();  // Annuler le comportement par défaut du lien
            const filmContent = this.parentElement;
            const overview = filmContent.querySelector(".film-overview");
            
            // Si la description est déjà étendue, la replier
            if (overview.classList.contains("expanded")) {
                overview.classList.remove("expanded");
                this.textContent = "En savoir plus";
            } else { // Sinon, l'étendre et modifier le texte du bouton
                overview.classList.add("expanded");
                this.textContent = "Réduire";
            }
        });
    });
});
