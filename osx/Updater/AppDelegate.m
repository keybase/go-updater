//
//  AppDelegate.m
//  Updater
//
//  Created by Gabriel on 4/7/16.
//  Copyright © 2016 Keybase. All rights reserved.
//

#import "AppDelegate.h"

#import "Defines.h"
#import "Prompt.h"

@implementation AppDelegate

- (void)applicationDidFinishLaunching:(NSNotification *)notification {
  dispatch_async(dispatch_get_main_queue(), ^{
    [self run];
  });
}

/*!
 If error specified, then exit with status 1, otherwise 0.
 */
- (void)exit:(NSError *)error {
  if (error) {
    NSLog(@"Error: %@", error.localizedDescription);
  }
  fflush(stdout);
  fflush(stderr);
  if (error) {
    exit(1);
  }
  exit(0);
}

- (void)run {
  NSArray *args = NSProcessInfo.processInfo.arguments;
  if (args.count < 1) {
    [self exit:KBMakeError(@"No args for process")];
  }

  NSArray *subargs = [args subarrayWithRange:NSMakeRange(1, args.count-1)];
  if (subargs.count < 1) {
    [self exit:KBMakeError(@"No args")];
  }

  [Prompt showPromptWithInputString:subargs[0] presenter:^NSModalResponse(NSAlert *alert) {
    return [alert runModal];
  } completion:^(NSError *error, NSData *output) {
    if (error) {
      [self exit:error];
      return;
    }
    [[NSFileHandle fileHandleWithStandardOutput] writeData:output];
    [self exit:nil];
  }];
}

@end
