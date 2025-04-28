from flask import Flask, request, Response
from curl_cffi import requests
import time

app = Flask(__name__)

@app.route('/', methods=['POST'])
def fetch_problem():
    url = request.data.decode().strip()
    print(f"Fetching URL: {url}")

    if not url.startswith('http'):
        return Response("Invalid URL", status=400)

    try:
        session = requests.Session(
            impersonate="chrome110"  # look like a real Chrome
        )

        # do initial get
        resp = session.get(url, timeout=60)

        if "prove you are human" in resp.text.lower():
            print("Detected Cloudflare challenge, retrying...")
            time.sleep(10)
            resp = session.get(url, timeout=60)

        from bs4 import BeautifulSoup
        soup = BeautifulSoup(resp.content, "html.parser")

        problem_statement = soup.select_one(".problem-statement")
        if problem_statement:
            return problem_statement.prettify()
        else:
            return Response("Problem statement not found.", status=404)

    except Exception as e:
        print("Error:", str(e))
        return Response(str(e), status=500)

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=3001)
