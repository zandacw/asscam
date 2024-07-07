webcam: test.c
	gcc -o webcam test.c `pkg-config --cflags --libs opencv`
