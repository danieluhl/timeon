#ifndef TIMEON_MACOS_H
#define TIMEON_MACOS_H

#include <stddef.h>

char *timeon_frontmost_app(void);
double timeon_idle_seconds(void);
int timeon_screen_locked(void);
int timeon_accessibility_trusted(void);
void timeon_request_accessibility(void);
void timeon_frontmost_diag(char *ax, char *ls, char *se, char *ws, size_t bufsize);

#endif
