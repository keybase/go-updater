//
//  AppDelegate.m
//  Updater
//
//  Created by Gabriel on 4/7/16.
//  Copyright © 2016 Gabriel Handford. All rights reserved.
//

#import "AppDelegate.h"

#import "Defines.h"
#import "Prompt.h"

@implementation AppDelegate

- (void)applicationDidFinishLaunching:(NSNotification *)notification {
#ifdef TEST
  return
#endif

  dispatch_async(dispatch_get_main_queue(), ^{
    [self run];
  });
}

- (void)exitWithError:(NSError *)error {
  NSLog(@"Error: %@", error.localizedDescription);
  fflush(stderr);
  exit(1);
}

- (void)run {
  NSArray *args = NSProcessInfo.processInfo.arguments;
  if (args.count < 1) {
    [self exitWithError:KBMakeError(@"No args for process")];
  }

  NSArray *subargs = [args subarrayWithRange:NSMakeRange(1, args.count-1)];
  if (subargs.count < 1) {
    [self exitWithError:KBMakeError(@"No args")];
  }

  if (![subargs[0] isKindOfClass:NSString.class]) {
    [self exitWithError:KBMakeError(@"Invalid arg")];
  }

  [Prompt showPromptWithInputString:subargs[0] presenter:^NSModalResponse(NSAlert *alert) {
    return [alert runModal];
  } completion:^(NSError *error, NSData *output) {
    if (error) {
      [self exitWithError:error];
    } else {
      [[NSFileHandle fileHandleWithStandardOutput] writeData:output];
      fflush(stdout);
      exit(0);
    }
  }];
}

@end
