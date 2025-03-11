<#import "template.ftl" as layout>
<#import "password-commons.ftl" as passwordCommons>
<@layout.registrationLayout displayMessage=!messagesPerField.existsError('password','password-confirm'); section>
    <#if section = "header">
        ${msg("updatePasswordTitle")}
    <#elseif section = "form">
        <form id="kc-passwd-update-form" class="${properties.kcFormClass!}" action="${url.loginAction}" method="post">
            <div class="${properties.kcFormGroupClass!}">
                <div class="${properties.kcLabelWrapperClass!}">
                    <label for="password-new" class="${properties.kcLabelClass!}">${msg("passwordNew")}</label>
                </div>
                <div class="${properties.kcInputWrapperClass!}">
                    <div class="pf-c-input-group" dir="ltr">
                            <div class="input-wrapper password-input">
                                <input 
                                    tabindex="3" 
                                    id="password" 
                                    class="${properties.kcInputClass!}" 
                                    name="password-new" 
                                    type="password" 
                                    autocomplete="current-password" 
                                    placeholder="Password" 
                                    aria-invalid=""
                                >
                                <button 
                                    class="pf-c-button" 
                                    type="button" 
                                    aria-label="Show password" 
                                    aria-controls="password" 
                                    data-password-toggle="" 
                                    tabindex="4" 
                                    data-icon-show="fa fa-eye" 
                                    data-icon-hide="fa fa-eye-slash" 
                                    data-label-show="Show password" 
                                    data-label-hide="Hide password"
                                >
                                    <i class="fa fa-eye" aria-hidden="true"></i>
                                </button>
                            </div>
                        </div>

                    <#if messagesPerField.existsError('password')>
                        <span id="input-error-password" class="${properties.kcInputErrorMessageClass!}" aria-live="polite">
                            ${kcSanitize(messagesPerField.get('password'))?no_esc}
                        </span>
                    </#if>
                </div>
            </div>

            <div class="${properties.kcFormGroupClass!}">
                <div class="${properties.kcLabelWrapperClass!}">
                    <label for="password-confirm" class="${properties.kcLabelClass!}">${msg("passwordConfirm")}</label>
                </div>
                <div class="${properties.kcInputWrapperClass!}">
                    <div class="pf-c-input-group" dir="ltr">
                            <div class="input-wrapper password-input">
                                <input 
                                    tabindex="3" 
                                    id="password-confirm" 
                                    class="${properties.kcInputClass!}" 
                                    name="password-confirm" 
                                    type="password" 
                                    autocomplete="current-password" 
                                    placeholder="Password" 
                                    aria-invalid=""
                                >
                                <button 
                                    class="pf-c-button" 
                                    type="button" 
                                    aria-label="Show password" 
                                    aria-controls="password-confirm" 
                                    data-password-toggle="" 
                                    tabindex="4" 
                                    data-icon-show="fa fa-eye" 
                                    data-icon-hide="fa fa-eye-slash" 
                                    data-label-show="Show password" 
                                    data-label-hide="Hide password"
                                >
                                    <i class="fa fa-eye" aria-hidden="true"></i>
                                </button>
                            </div>
                        </div>

                    <#if messagesPerField.existsError('password-confirm')>
                        <span id="input-error-password-confirm" class="${properties.kcInputErrorMessageClass!}" aria-live="polite">
                            ${kcSanitize(messagesPerField.get('password-confirm'))?no_esc}
                        </span>
                    </#if>

                </div>
            </div>

            <div class="${properties.kcFormGroupClass!}">
                <@passwordCommons.logoutOtherSessions/>

                <div id="kc-form-buttons" class="${properties.kcFormButtonsClass!}">
                    <#if isAppInitiatedAction??>
                        <input class="${properties.kcButtonClass!} ${properties.kcButtonPrimaryClass!} ${properties.kcButtonLargeClass!}" type="submit" value="${msg("doSubmit")}" />
                        <button class="${properties.kcButtonClass!} ${properties.kcButtonDefaultClass!} ${properties.kcButtonLargeClass!}" type="submit" name="cancel-aia" value="true" />${msg("doCancel")}</button>
                    <#else>
                        <input class="${properties.kcButtonClass!} ${properties.kcButtonPrimaryClass!} ${properties.kcButtonBlockClass!} ${properties.kcButtonLargeClass!}" type="submit" value="${msg("doSubmit")}" />
                    </#if>
                </div>
            </div>
        </form>
        <script type="module" src="${url.resourcesPath}/js/passwordVisibility.js"></script>
    </#if>
</@layout.registrationLayout>
