document.getElementById('placeOrderForm').addEventListener('submit', function(event) {
    event.preventDefault();
    const formData = new FormData(event.target);
    const orderData = {
        type: formData.get('type'),
        action: formData.get('action'),
        size: parseFloat(formData.get('size')),
        price: parseFloat(formData.get('price')),
        ticker: formData.get('ticker'),
    };
    fetch('/order', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(orderData),
    })
    .then(response => response.json())
    .then(data => {
        alert(JSON.stringify(data));
        document.getElementById('refreshOrderBook').click();
    })
    .catch(error => console.error('Error placing order:', error));
});

document.getElementById('cancelOrderForm').addEventListener('submit', function(event) {
    event.preventDefault();
    const orderId = new FormData(event.target).get('id');
    fetch(`/order/${orderId}`, {
        method: 'DELETE',
    })
    .then(response => response.json())
    .then(data => {
        alert(JSON.stringify(data));
        document.getElementById('refreshOrderBook').click();
    })
    .catch(error => console.error('Error canceling order:', error));
});

document.getElementById('refreshOrderBook').addEventListener('click', function() {
    fetch('/book/ETH')
    .then(response => response.json())
    .then(data => {
        document.getElementById('orderBook').textContent = JSON.stringify(data, null, 2);
    })
    .catch(error => console.error('Error fetching order book:', error));
});

document.getElementById('refreshOrderBook').click();
