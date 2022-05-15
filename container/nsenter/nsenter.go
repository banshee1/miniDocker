package nsenter

/*
#define _GNU_SOURCE
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h>

__attribute__((constructor)) void enter_namespace(void) {
	char *mydocker_pid;
	mydocker_pid = getenv("mdocker_pid");
	if (mydocker_pid) {
		fprintf(stdout, "got mdocker_pid=%s\n", mydocker_pid);
	} else {
		fprintf(stdout, "missing mdocker_pid env skip nsenter\n");
		return;
	}

	char *mydocker_cmd;
	mydocker_cmd = getenv("mdocker_cmd");
	if (mydocker_cmd) {
		fprintf(stdout, "got mdocker_cmd=%s\n", mydocker_cmd);
	} else {
		fprintf(stdout, "missing mdocker_cmd env skip nsenter\n");
		return;
	}
	int i;
	char nspath[1024];
	char *namespaces[] = { "ipc", "uts", "net", "pid", "mnt" };

	for (i=0; i<5; i++) {
		sprintf(nspath, "/proc/%s/ns/%s", mydocker_pid, namespaces[i]);
		int fd = open(nspath, O_RDONLY);

		if (setns(fd, 0) == -1) {
			fprintf(stderr, "setns on %s namespace failed: %s\n", namespaces[i], strerror(errno));
		} else {
			fprintf(stdout, "setns on %s namespace succeeded\n", namespaces[i]);
		}
		close(fd);
	}
	int res = system(mydocker_cmd);
	exit(0);
	return;
}
*/
import "C"