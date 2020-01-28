window.onload = () => {
    getPosts()
}

function submitPost(e) {
    e.preventDefault();

    const name = document.getElementById("postName").value;
    const homepage = document.getElementById("postHomepage").value;
    const comment = document.getElementById("postComment").value;
    const email = document.getElementById("postEmail").value;

    document.getElementById("postForm").reset();

    const params = {
        method: "POST",
        body: JSON.stringify({name, homepage, comment, email})
    }
    const url = `http://localhost:8080/post`
    fetch(url, params)
        .then(res => {
            console.log(res)
            getPosts()
        })
        .catch(err => console.log(err))
}

function getPosts() {
    const params = {
        method: "GET"
    }
    const url = `http://localhost:8080/post`
    fetch(url, params)
        .then(res => res.json())
        .then(posts => {
            console.log(posts)
            var postCol = document.getElementById("postsCol");
            if(posts.length === 0) {
                postCol.innerHTML = "No posts yet..."
                return
            }
            postCol.innerHTML = ""
            posts.forEach(post => {
                postCol.innerHTML += `
                <div class="row border">
                    <div class="col">
                        <div class="row">
                            <div class="col-4"><b>Name:</b> ${post.name}</div>
                            <div class="col-4"><b>Email:</b> ${post.email}</div>
                            <div class="col-4"><b>Homepage:</b> ${post.homepage}</div>
                        </div>
                        <div class="row mt-2">
                            <div class="col">
                                <b>Comment:</b> ${post.comment}
                            </div>
                        </div>
                    </div>
                </div>
                `
            })
        })
        .catch(err => console.log(err))

}