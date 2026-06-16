#import <AppKit/AppKit.h>
#import <ApplicationServices/ApplicationServices.h>
#import <CoreGraphics/CoreGraphics.h>
#import <stdio.h>
#import <stdlib.h>
#import <string.h>
#import "macos.h"

static char *copy_nsstring(NSString *value) {
    if (value == nil || value.length == 0) {
        return NULL;
    }
    return strdup([value UTF8String]);
}

static char *parse_lsdisplay_name(const char *line) {
    if (line == NULL) {
        return NULL;
    }

    const char *start = strstr(line, "=\"");
    if (start != NULL) {
        start += 2;
        const char *end = strchr(start, '"');
        if (end == NULL || end == start) {
            return NULL;
        }
        size_t len = (size_t)(end - start);
        char *out = malloc(len + 1);
        if (out == NULL) {
            return NULL;
        }
        memcpy(out, start, len);
        out[len] = '\0';
        return out;
    }

    size_t len = strlen(line);
    while (len > 0 && (line[len - 1] == '\n' || line[len - 1] == '\r')) {
        len--;
    }
    if (len == 0) {
        return NULL;
    }
    char *out = malloc(len + 1);
    if (out == NULL) {
        return NULL;
    }
    memcpy(out, line, len);
    out[len] = '\0';
    return out;
}

static void copy_into_buffer(char *dest, size_t bufsize, const char *value) {
    if (dest == NULL || bufsize == 0) {
        return;
    }
    if (value == NULL) {
        snprintf(dest, bufsize, "(unavailable)");
        return;
    }
    snprintf(dest, bufsize, "%s", value);
}

static char *app_name_for_pid(pid_t pid) {
    NSRunningApplication *app = [NSRunningApplication runningApplicationWithProcessIdentifier:pid];
    if (app == nil) {
        return NULL;
    }

    NSString *name = app.localizedName;
    if (name == nil || name.length == 0) {
        NSString *bundleID = app.bundleIdentifier;
        if (bundleID != nil) {
            name = bundleID.lastPathComponent;
        }
    }
    return copy_nsstring(name);
}

static char *frontmost_from_lsappinfo(void) {
    FILE *fp = popen(
        "/bin/sh -c '/usr/bin/lsappinfo info -only name \"$(/usr/bin/lsappinfo front)\" 2>/dev/null'",
        "r"
    );
    if (fp == NULL) {
        return NULL;
    }

    char line[256];
    if (fgets(line, sizeof(line), fp) == NULL) {
        pclose(fp);
        return NULL;
    }
    pclose(fp);
    return parse_lsdisplay_name(line);
}

static char *frontmost_from_ax(void) {
    if (!AXIsProcessTrusted()) {
        return NULL;
    }

    AXUIElementRef system = AXUIElementCreateSystemWide();
    if (system == NULL) {
        return NULL;
    }

    CFTypeRef focused = NULL;
    AXError err = AXUIElementCopyAttributeValue(
        system,
        kAXFocusedApplicationAttribute,
        &focused
    );
    CFRelease(system);

    if (err != kAXErrorSuccess || focused == NULL) {
        return NULL;
    }

    pid_t pid = 0;
    err = AXUIElementGetPid((AXUIElementRef)focused, &pid);
    CFRelease(focused);

    if (err != kAXErrorSuccess || pid <= 0) {
        return NULL;
    }

    return app_name_for_pid(pid);
}

static char *frontmost_from_system_events(void) {
    NSAppleScript *script = [[NSAppleScript alloc] initWithSource:
        @"tell application \"System Events\" to get name of first application process whose frontmost is true"];
    NSDictionary *errorInfo = nil;
    NSAppleEventDescriptor *result = [script executeAndReturnError:&errorInfo];
    if (result == nil) {
        return NULL;
    }
    return copy_nsstring([result stringValue]);
}

static char *frontmost_from_workspace(void) {
    NSRunningApplication *app = [[NSWorkspace sharedWorkspace] frontmostApplication];
    if (app == nil) {
        return NULL;
    }
    return copy_nsstring(app.localizedName);
}

char *timeon_frontmost_app(void) {
    @autoreleasepool {
        char *name = frontmost_from_lsappinfo();
        if (name != NULL) {
            return name;
        }

        name = frontmost_from_ax();
        if (name != NULL) {
            return name;
        }

        name = frontmost_from_system_events();
        if (name != NULL) {
            return name;
        }

        name = frontmost_from_workspace();
        if (name != NULL) {
            return name;
        }

        return strdup("Unknown");
    }
}

void timeon_frontmost_diag(char *ax, char *ls, char *se, char *ws, size_t bufsize) {
    @autoreleasepool {
        char *axName = frontmost_from_ax();
        char *lsName = frontmost_from_lsappinfo();
        char *seName = frontmost_from_system_events();
        char *wsName = frontmost_from_workspace();

        copy_into_buffer(ax, bufsize, axName);
        copy_into_buffer(ls, bufsize, lsName);
        copy_into_buffer(se, bufsize, seName);
        copy_into_buffer(ws, bufsize, wsName);

        free(axName);
        free(lsName);
        free(seName);
        free(wsName);
    }
}

int timeon_accessibility_trusted(void) {
    return AXIsProcessTrusted() ? 1 : 0;
}

void timeon_request_accessibility(void) {
    NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @YES};
    AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
}

double timeon_idle_seconds(void) {
    return CGEventSourceSecondsSinceLastEventType(
        kCGEventSourceStateHIDSystemState,
        kCGAnyInputEventType
    );
}

int timeon_screen_locked(void) {
    CFDictionaryRef session = CGSessionCopyCurrentDictionary();
    if (session == NULL) {
        return 0;
    }

    const void *value = NULL;
    int locked = 0;
    if (CFDictionaryGetValueIfPresent(session, CFSTR("CGSSessionScreenIsLocked"), &value)) {
        if (value != NULL && CFGetTypeID(value) == CFBooleanGetTypeID()) {
            locked = CFBooleanGetValue((CFBooleanRef)value) ? 1 : 0;
        }
    }

    CFRelease(session);
    return locked;
}
