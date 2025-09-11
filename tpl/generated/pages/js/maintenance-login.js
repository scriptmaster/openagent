
    // Load a random tech GIF from Tenor
    document.addEventListener('DOMContentLoaded', function() {
        // Tenor GIF API
        fetch('https://g.tenor.com/v1/random?q=tech+funny&key=LIVDSRZULELA&limit=1')
            .then(response => response.json())
            .then(data => {
                if (data.results && data.results.length > 0) {
                    document.querySelector('.trending-gif').src = data.results[0].media[0].gif.url;
                }
            })
            .catch(error => {
                console.error('Error fetching GIF:', error);
                // Keep the default GIF if there's an error
            });
    });

