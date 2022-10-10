import cv2
import os
import time
from dotenv import load_dotenv

load_dotenv()

IMG_FOLDER = os.getenv("IMG_FOLDER")
VIDEO_FILE = os.getenv("VIDEO_FILE")
TIME_FRAME = float(os.getenv("TIME_FRAME"))
TIME_WAIT = int(os.getenv("TIME_WAIT"))

images = [img for img in os.listdir(IMG_FOLDER) if img.endswith(".png")]
frame = cv2.imread(os.path.join(IMG_FOLDER, images[0]))
height, width, layers = frame.shape
video = cv2.VideoWriter(VIDEO_FILE, 0, 5, (width, height))

for image in images:
    video.write(cv2.imread(os.path.join(IMG_FOLDER, image)))
cv2.destroyAllWindows()
video.release()
cap = cv2.VideoCapture(VIDEO_FILE)
cv2.namedWindow('kmeans', cv2.WINDOW_AUTOSIZE)

while True:
    try:
        ret_val, frame = cap.read()
        time.sleep(TIME_FRAME)
        cv2.imshow('kmeans', frame)
        if cv2.waitKey(1) == 27:
            break
    except:
        break
time.sleep(TIME_WAIT)
