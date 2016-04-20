//
//  PromptTests.m
//  UpdaterTests
//
//  Created by Gabriel on 4/13/16.
//  Copyright © 2016 Gabriel Handford. All rights reserved.
//

#import <XCTest/XCTest.h>

#import "Prompt.h"

@interface PromptTests : XCTestCase
@end

@implementation PromptTests

- (void)testUpdatePrompt {
  NSData *data = [NSJSONSerialization dataWithJSONObject:@{@"title": @"Title", @"message": @"", @"description": @"", @"autoUpdate": @NO} options:0 error:nil];
  NSString *str = [[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding];
  [Prompt showPromptWithInputString:str presenter:^NSModalResponse(NSAlert *alert) {
    return NSAlertFirstButtonReturn;
  } completion:^(NSError *error, NSData *output) {
    NSLog(@"Error: %@", error);
    NSLog(@"Output: %@", output);
    XCTAssertNil(error);
  }];
}

- (void)testUpdatePromptNoSettings {
  [Prompt showPromptWithInputString:@"" presenter:^NSModalResponse(NSAlert *alert) {
    return NSAlertFirstButtonReturn;
  } completion:^(NSError *error, NSData *output) {
    NSLog(@"Error: %@", error);
    NSLog(@"Output: %@", output);
    XCTAssertNil(error);
  }];
}

- (void)testUpdatePromptInvalidJSONInput {
  NSData *data = [NSJSONSerialization dataWithJSONObject:@{@"title": @{}, @"message": @{}, @"description": @{}, @"autoUpdate": @{}} options:0 error:nil];
  NSString *str = [[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding];
  [Prompt showPromptWithInputString:str presenter:^NSModalResponse(NSAlert *alert) {
    return NSAlertFirstButtonReturn;
  } completion:^(NSError *error, NSData *output) {
    NSLog(@"Error: %@", error);
    NSLog(@"Output: %@", output);
    XCTAssertNil(error);
  }];
}

- (void)testUpdatePromptInvalidJSONRoot {
  NSData *data = [NSJSONSerialization dataWithJSONObject:@[] options:0 error:nil];
  NSString *str = [[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding];
  [Prompt showPromptWithInputString:str presenter:^NSModalResponse(NSAlert *alert) {
    return NSAlertFirstButtonReturn;
  } completion:^(NSError *error, NSData *output) {
    NSLog(@"Error: %@", error);
    NSLog(@"Output: %@", output);
    XCTAssertNil(error);
  }];
}

@end
