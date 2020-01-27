function sendEmail(e) {
    e.preventDefault();
    // Email server data
    const server = document.getElementById("emailServer").value;
    const port = document.getElementById("emailPort").value;
    const user = document.getElementById("emailServerUsername").value;
    const password = document.getElementById("emailServerPassword").value;

    // Email data
    const from = document.getElementById("emailFrom").value;
    const to = document.getElementById("emailTo").value;
    const subject = document.getElementById("emailSubject").value;
    const message = document.getElementById("emailMessage").value;

    document.getElementById("emailForm").reset();
    const params = {
        method: "POST",
        body: JSON.stringify({
            "server": server,
            "port": port,
            "user": user,
            "password": password,
            "to": to,
            "from": from,
            "subject": subject,
            "message": message,
        })
    }
    console.log(params)
    const url = `http://localhost:8080/email`
    fetch(url, params)
        .then(res => console.log(res))
        .catch(err => console.log(err))

}