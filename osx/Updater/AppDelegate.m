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
  // Check if test environment
  if ([self isRunningTests]) return;

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
  NSString *inputString = @"{}";

  NSArray *args = NSProcessInfo.processInfo.arguments;
  if (args.count > 0) {
    NSArray *subargs = [args subarrayWithRange:NSMakeRange(1, args.count-1)];
    if (subargs.count >= 1) {
      inputString = subargs[0];
    }
  }

  [Prompt showPromptWithInputString:inputString presenter:^NSModalResponse(NSAlert *alert) {
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

- (BOOL)isRunningTests {
  // The Xcode test environment is a little awkward. Instead of using TEST preprocessor macro, check env.
  NSDictionary *environment = [[NSProcessInfo processInfo] environment];
  NSString *testFilePath = environment[@"XCTestConfigurationFilePath"];
  return !!testFilePath;
}

@end
