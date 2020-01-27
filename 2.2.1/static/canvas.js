window.onload = function() {
    const canvas = document.getElementById('draw_canvas');
    const ctx = canvas.getContext("2d")
    resizeCanvas(ctx);
    const bounds = canvas.getBoundingClientRect();

    // Create websocket
    const socket = new WebSocket(`ws://localhost:${srvPort}/ws`);

    let isDrawing = false;
    let line = []
    let xPrev = 0, yPrev = 0;

    const url = `http://localhost:${srvPort}/lines`
    fetch(url)
        .then(res => res.json())
        .then(json => {
            json.forEach(line => {
                const cords = line.coords
                let prev = cords.shift()
                cords.forEach((next) => {
                    drawLine(ctx, prev.x, prev.y, next.x, next.y)
                    prev = next
                })
            })
        })
        .catch(err => this.console.log(err))

    // Mouse events for line drawing
    canvas.onmousedown = (e) => {
        isDrawing = true;
        xPrev = e.clientX - bounds.left, yPrev = e.clientY - bounds.top;
        line.push({x: xPrev, y: yPrev});
    };

    canvas.onmousemove = (e) => {
        if(isDrawing) {
            let x = e.clientX - bounds.left, y = e.clientY - bounds.top;
            drawLine(ctx, xPrev, yPrev, x, y);
            xPrev = x, yPrev = y;
            line.push({x: xPrev, y: yPrev});
        }
    };

    canvas.onmouseup = (e) => {
        if(isDrawing) {
            let x = e.clientX - bounds.left, y = e.clientY - bounds.top;
            line.push({x: x, y: y});
            drawLine(ctx, xPrev, yPrev, x, y);
            isDrawing = false;
            // Send new line to server, which will distribute the line to peers
            socket.send(JSON.stringify(line));
            line = []
        }
    };
    
    // Draw any lines recieved on socket on canvas
    socket.onmessage = (event) => {
        let line = JSON.parse(event.data)
        let prev = line.shift()
        line.forEach((next) => {
            drawLine(ctx, prev.x, prev.y, next.x, next.y)
            prev = next
        })
    }
};

function resizeCanvas(ctx) {
    ctx.canvas.width  = window.innerWidth * 0.75;
    ctx.canvas.height = window.innerHeight * 0.75;
}

function drawLine(ctx, x1, y1, x2, y2) {
    ctx.beginPath();
    ctx.strokeStyle = 'black';
    ctx.lineWidth = 1;
    ctx.moveTo(x1, y1);
    ctx.lineTo(x2, y2);
    ctx.stroke();
    ctx.closePath();
}