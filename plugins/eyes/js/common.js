var Common = {
    confirm:function(params){
        var model = $("#common_confirm_model");
        model.find(".title").html(params.title)
        model.find(".message").html(params.message)
        model.find(".cancel").unbind();
        model.find(".ok").unbind();
        model.find(".ok").click(function(){
            params.operate(true)
        })
        model.find(".cancel").click(function(){
            params.operate(false)
        })
        model.modal({show: true,backdrop: 'static'});
    },
    info:function(params){
        var model = $("#myModal");
        model.find(".title").html(params.title)
        model.find(".message").html(params.message)
        model.find(".ok").unbind();
        model.find(".ok").click(function(){
            params.operate(true)
        })
        model.modal({show: true,backdrop: 'static'});
    }

}

