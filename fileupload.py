import requests

url = 'http://localhost:8080/uploadfile'
files = {'file': open('/home/samvit.swaminathan/Desktop/go-prac/image1.jpg', 'rb')}
response = requests.post(url, files=files)
print(response.text)
