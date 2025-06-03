// Animation d'entrée progressive
        document.addEventListener('DOMContentLoaded', function() {
            // Animer le logo
            const brandSection = document.querySelector('.brand-section');
            brandSection.style.opacity = '0';
            brandSection.style.transform = 'translateY(-30px)';
            
            setTimeout(() => {
                brandSection.style.transition = 'opacity 1s ease, transform 1s ease';
                brandSection.style.opacity = '1';
                brandSection.style.transform = 'translateY(0)';
            }, 200);

            // Effet de particules au survol des cartes
            document.querySelectorAll('.page-card').forEach(card => {
                card.addEventListener('mouseenter', function() {
                    this.style.boxShadow = '0 25px 50px rgba(102, 126, 234, 0.2)';
                });
                
                card.addEventListener('mouseleave', function() {
                    this.style.boxShadow = '0 20px 40px rgba(0, 0, 0, 0.2)';
                });
            });

            // Message de bienvenue
            setTimeout(() => {
                console.log('🎵 Bienvenue sur Rythm\'it !');
                console.log('🚀 Explorez toutes les fonctionnalités du réseau social musical');
                console.log('📖 Consultez le README.md pour plus d\'informations');
            }, 2000);
        });