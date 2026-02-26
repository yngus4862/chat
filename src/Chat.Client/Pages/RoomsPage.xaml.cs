using Chat.Client.ViewModels;

namespace Chat.Client.Pages;

public partial class RoomsPage : ContentPage
{
    public RoomsPage(RoomsViewModel vm)
    {
        InitializeComponent();
        BindingContext = vm;
    }

    protected override async void OnAppearing()
    {
        base.OnAppearing();
        if (BindingContext is RoomsViewModel vm)
            await vm.EnsureLoadedAsync();
    }
}