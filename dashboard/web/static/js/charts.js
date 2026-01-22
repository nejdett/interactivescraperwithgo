// Charts and Statistics JavaScript

let categoryChart = null;
let criticalityChart = null;
let statsRefreshInterval = null;

// Initialize charts when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    if (window.location.pathname === '/dashboard') {
        initializeCharts();
        loadStatistics();
        
        // Auto-refresh every 30 seconds
        statsRefreshInterval = setInterval(loadStatistics, 30000);
    }
});

// Clean up interval when leaving page
window.addEventListener('beforeunload', function() {
    if (statsRefreshInterval) {
        clearInterval(statsRefreshInterval);
    }
});

async function loadStatistics() {
    try {
        const response = await fetch('/api/contents/stats', {
            credentials: 'include'
        });
        
        if (!response.ok) {
            throw new Error('Failed to load statistics');
        }
        
        const stats = await response.json();
        updateStatisticsCards(stats);
        updateCharts(stats);
    } catch (error) {
        console.error('Error loading statistics:', error);
    }
}

function updateStatisticsCards(stats) {
    // Update total items
    document.getElementById('stat-total').textContent = stats.total_items || 0;
    
    // Calculate high criticality count (9-10)
    const highCriticalityCount = stats.criticality_distribution['9-10'] || 0;
    document.getElementById('stat-high-criticality').textContent = highCriticalityCount;
    
    // Update last updated time
    if (stats.last_updated) {
        const lastUpdated = new Date(stats.last_updated);
        const timeString = lastUpdated.toLocaleTimeString('en-US', { 
            hour: '2-digit', 
            minute: '2-digit' 
        });
        document.getElementById('stat-last-updated').textContent = timeString;
    }
}

function initializeCharts() {
    // Initialize Category Distribution Pie Chart
    const categoryCtx = document.getElementById('category-chart');
    if (categoryCtx) {
        categoryChart = new Chart(categoryCtx, {
            type: 'pie',
            data: {
                labels: [],
                datasets: [{
                    data: [],
                    backgroundColor: [
                        '#e74c3c', // Red
                        '#f39c12', // Orange
                        '#3498db', // Blue
                        '#9b59b6', // Purple
                        '#1abc9c', // Teal
                        '#34495e', // Dark gray
                        '#e67e22', // Dark orange
                        '#2ecc71'  // Green
                    ],
                    borderWidth: 2,
                    borderColor: '#fff'
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'bottom',
                        labels: {
                            padding: 15,
                            font: {
                                size: 12
                            }
                        }
                    },
                    tooltip: {
                        callbacks: {
                            label: function(context) {
                                const label = context.label || '';
                                const value = context.parsed || 0;
                                const total = context.dataset.data.reduce((a, b) => a + b, 0);
                                const percentage = ((value / total) * 100).toFixed(1);
                                return `${label}: ${value} (${percentage}%)`;
                            }
                        }
                    }
                }
            }
        });
    }
    
    // Initialize Criticality Distribution Bar Chart
    const criticalityCtx = document.getElementById('criticality-chart');
    if (criticalityCtx) {
        criticalityChart = new Chart(criticalityCtx, {
            type: 'bar',
            data: {
                labels: ['Low (1-3)', 'Medium (4-6)', 'High (7-8)', 'Critical (9-10)'],
                datasets: [{
                    label: 'Number of Items',
                    data: [0, 0, 0, 0],
                    backgroundColor: [
                        '#95a5a6', // Gray for low
                        '#3498db', // Blue for medium
                        '#f39c12', // Orange for high
                        '#e74c3c'  // Red for critical
                    ],
                    borderWidth: 0
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: false
                    },
                    tooltip: {
                        callbacks: {
                            title: function(context) {
                                return context[0].label;
                            },
                            label: function(context) {
                                const value = context.parsed.y || 0;
                                return `Items: ${value}`;
                            },
                            afterLabel: function(context) {
                                const labels = ['Low priority threats', 'Medium priority threats', 'High priority - requires attention', 'Critical - immediate action required'];
                                return labels[context.dataIndex];
                            }
                        }
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        ticks: {
                            stepSize: 1
                        }
                    }
                }
            }
        });
    }
}

function updateCharts(stats) {
    // Update Category Chart
    if (categoryChart && stats.category_distribution) {
        const categories = Object.keys(stats.category_distribution);
        const values = Object.values(stats.category_distribution);
        
        categoryChart.data.labels = categories;
        categoryChart.data.datasets[0].data = values;
        categoryChart.update();
    }
    
    // Update Criticality Chart
    if (criticalityChart && stats.criticality_distribution) {
        const criticalityData = [
            stats.criticality_distribution['1-3'] || 0,
            stats.criticality_distribution['4-6'] || 0,
            stats.criticality_distribution['7-8'] || 0,
            stats.criticality_distribution['9-10'] || 0
        ];
        
        criticalityChart.data.datasets[0].data = criticalityData;
        criticalityChart.update();
    }
}
