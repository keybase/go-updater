//
//  Prompt.m
//  Updater
//
//  Created by Gabriel on 4/13/16.
//  Copyright © 2016 Keybase. All rights reserved.
//

#import "Prompt.h"
#import "TextView.h"
#import "Defines.h"
#import "NSDictionary+Extension.h"

@interface FView : NSView
@end

@implementation Prompt

+ (NSDictionary *)parseInputString:(NSString *)inputString defaultValue:(NSDictionary *)defaultValue {
  NSData *data = [inputString dataUsingEncoding:NSUTF8StringEncoding];
  if (!data) {
    NSLog(@"No data for input");
    return defaultValue;
  }
  NSError *error = nil;
  id input = [NSJSONSerialization JSONObjectWithData:data options:0 error:&error];
  if (!!error) {
    NSLog(@"Error parsing input: %@", error);
    return defaultValue;
  }
  if (!input) {
    NSLog(@"No input");
    return defaultValue;
  }
  if (![input isKindOfClass:[NSDictionary class]]) {
    NSLog(@"Invalid input type");
    return defaultValue;
  }
  return input;
}

+ (void)showPromptWithInputString:(NSString *)inputString presenter:(NSModalResponse (^)(NSAlert *alert))presenter completion:(void (^)(NSError *error, NSData *output))completion {

  // Try to parse input, if there is any error use a default empty dictionary.
  NSDictionary *input = [self parseInputString:inputString defaultValue:@{}];

  [self showUpdatePrompt:input presenter:presenter completion:completion];
}

+ (void)showUpdatePrompt:(NSDictionary *)input presenter:(NSModalResponse (^)(NSAlert *alert))presenter completion:(void (^)(NSError *error, NSData *output))completion {
  NSString *title = [input kb_stringForKey:@"title"];
  NSString *message = [input kb_stringForKey:@"message"];
  NSString *description = [input kb_stringForKey:@"description"];
  BOOL autoUpdate = [input kb_boolForKey:@"autoUpdate"];

  if (!title) title = @"Keybase Update";
  if (!message) message = @"There is an update available.";
  if (!description) description = @"Please visit keybase.io for more information.";

  NSAlert *alert = [[NSAlert alloc] init];
  alert.messageText = title;
  alert.informativeText = message;
  [alert addButtonWithTitle:@"Update"];
  [alert addButtonWithTitle:@"Ignore"];

  FView *accessoryView = [[FView alloc] init];
  TextView *textView = [[TextView alloc] init];
  textView.editable = NO;
  textView.view.textContainerInset = CGSizeMake(5, 5);

  NSFont *font = [NSFont fontWithName:@"Monaco" size:10];
  [textView setText:description font:font color:[NSColor blackColor] alignment:NSLeftTextAlignment lineBreakMode:NSLineBreakByCharWrapping];
  textView.borderType = NSBezelBorder;
  textView.frame = CGRectMake(0, 0, 500, 160);
  [accessoryView addSubview:textView];

  NSButton *autoCheckbox = [[NSButton alloc] init];
  autoCheckbox.title = @"Update automatically";
  autoCheckbox.state = autoUpdate ? NSOnState : NSOffState;
  [autoCheckbox setButtonType:NSSwitchButton];
  autoCheckbox.frame = CGRectMake(0, 160, 500, 30);
  [accessoryView addSubview:autoCheckbox];
  accessoryView.frame = CGRectMake(0, 0, 500, 190);
  alert.accessoryView = accessoryView;

  [alert setAlertStyle:NSInformationalAlertStyle];


  NSModalResponse response = presenter(alert);
  NSString *action = @"";
  if (response == NSAlertFirstButtonReturn) {
    action = @"update";
  } else if (response == NSAlertSecondButtonReturn) {
    action = @"ignore";
  }
  NSLog(@"Action: %@", action);

  NSError *error = nil;
  NSDictionary *result = @{
                           @"action": action,
                           @"autoUpdate": autoCheckbox.state == NSOnState ? @YES : @NO,
                           };

  NSData *data = [NSJSONSerialization dataWithJSONObject:result options:0 error:&error];
  if (!!error) {
    completion(error, nil);
    return;
  }
  if (!data) {
    completion(KBMakeError(@"No data"), nil);
    return;
  }
  completion(nil, data);
}

@end

@implementation FView

- (BOOL)isFlipped {
  return YES;
}

@end
