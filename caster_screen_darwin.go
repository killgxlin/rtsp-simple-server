package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework CoreGraphics
#import <Foundation/Foundation.h>
#include <sys/sysctl.h>
static bool _needGainPrivacy = false;
bool needGainPrivacy() {
    return _needGainPrivacy;
}
BOOL canRecordScreen()
{
    if (@available(macOS 10.15, *)) {
        CFArrayRef windowList = CGWindowListCopyWindowInfo(kCGWindowListOptionOnScreenOnly, kCGNullWindowID);
        NSUInteger numberOfWindows = CFArrayGetCount(windowList);
        NSUInteger numberOfWindowsWithInfoGet = 0;
        for (int idx = 0; idx < numberOfWindows; idx++) {
            NSDictionary *windowInfo = (NSDictionary *)CFArrayGetValueAtIndex(windowList, idx);
            NSString *windowName = windowInfo[(id)kCGWindowName];
            NSNumber* sharingType = windowInfo[(id)kCGWindowSharingState];
            if (windowName || kCGWindowSharingNone != sharingType.intValue) {
                NSNumber* pid = windowInfo[(id)kCGWindowOwnerPID];
                NSString* appName = windowInfo[(id)kCGWindowOwnerName];
                NSLog(@"windowInfo get success pid:%lu appName:%@", pid.integerValue, appName);
                numberOfWindowsWithInfoGet++;
            } else {
                NSNumber* pid = windowInfo[(id)kCGWindowOwnerPID];
                NSString* appName = windowInfo[(id)kCGWindowOwnerName];
                NSLog(@"windowInfo get Fail pid:%lu appName:%@", pid.integerValue, appName);
            }
        }
        NSLog(@"numberOfWindows:%lu numberOfWindowsWithInfoGet:%lu", numberOfWindows, numberOfWindowsWithInfoGet);
        CFRelease(windowList);
        if (numberOfWindows == numberOfWindowsWithInfoGet) {
            return YES;
        } else {
            return NO;
        }
    }
    return YES;
}

NSString * runCommand(NSString *commandToRun)
{
    NSTask *task = [[NSTask alloc] init];
    [task setLaunchPath:@"/bin/sh"];
    NSArray *arguments = [NSArray arrayWithObjects:
                          @"-c" ,
                          [NSString stringWithFormat:@"%@", commandToRun],
                          nil];
    NSLog(@"run command:%@", commandToRun);
    [task setArguments:arguments];
    NSPipe *pipe = [NSPipe pipe];
    [task setStandardOutput:pipe];
    NSFileHandle *file = [pipe fileHandleForReading];
    [task launch];
    NSData *data = [file readDataToEndOfFile];
    NSString *output = [[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding];
    return output;
}
// return true if canRecord
void preparePrivacy() {
    NSLog(@"preparePrivacy");
    if (!canRecordScreen()) {
        NSLog(@"canRecordScreen");
        _needGainPrivacy = true;
        return;
    }
    _needGainPrivacy = false;
}
*/
import "C"

func preparePrivacy() {
	C.preparePrivacy()
}

func needGainPrivacy() bool {
	return bool(C.needGainPrivacy())
}
